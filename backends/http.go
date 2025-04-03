package backends

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	envparse "github.com/hashicorp/go-envparse"
	"github.com/situation-sh/situation/models"
)

const (
	POST = http.MethodPost
	PUT  = http.MethodPut
)

// These variables are constant bytes within the code to "easily" modify
// default values without recompiling the binary.
// Yes, this is hacky but it does the job
var (
	defaultAuthorizationHeader = [...]byte{
		65, 103, 101, 110, 116, 32, 60, 80,
		82, 69, 70, 73, 88, 62, 46, 60,
		84, 72, 73, 82, 84, 89, 84, 87,
		79, 51, 50, 67, 72, 65, 82, 65,
		67, 84, 69, 82, 83, 70, 79, 82,
		65, 80, 73, 75, 69, 89, 62} // "Agent <PREFIX>.<THIRTYTWO32CHARACTERSFORAPIKEY>"
	defaultAuthorizationHeaderSHA256 = [...]byte{
		109, 168, 248, 49, 211, 64, 71, 251,
		169, 149, 183, 117, 76, 157, 127, 79,
		20, 133, 179, 136, 205, 135, 135, 16,
		170, 177, 28, 123, 9, 97, 47, 146} // sha256("Agent <PREFIX>.<THIRTYTWO32CHARACTERSFORAPIKEY>")
	defaultHttpBackendURLBytes = [256]byte{
		104, 116, 116, 112, 58, 47, 47, 108, 111, 99,
		97, 108, 104, 111, 115, 116, 58, 56, 48, 48,
		48, 47, 105, 109, 112, 111, 114, 116, 47, 115,
		105, 116, 117, 97, 116, 105, 111, 110, 47} // "http://localhost:8000/import/situation/"
)

var (
	contentTypeHeaderKey   = http.CanonicalHeaderKey("Content-Type")
	authorizationHeaderKey = http.CanonicalHeaderKey("Authorization")
)

type HttpBackend struct {
	url    string
	method string
	header http.Header
}

// default config
var defaultHttpBackend = HttpBackend{
	url:    getDefaultURL(),
	method: POST,
	header: http.Header{
		contentTypeHeaderKey:   []string{"application/json"},
		authorizationHeaderKey: getDefaultAuthorizationHeader(),
	},
}

func getDefaultAuthorizationHeader() []string {
	// compare with hash to avoid having the string in two places
	// inside the code
	sum := sha256.Sum256(defaultAuthorizationHeader[:])
	// fmt.Println(sum, defaultAuthorizationHeaderSHA256)
	if sum == defaultAuthorizationHeaderSHA256 {
		return nil
	}
	return []string{string(defaultAuthorizationHeader[:])}
}

func getDefaultURL() string {
	// stop to the first null byte
	index := 0
	for ; defaultHttpBackendURLBytes[index] != 0; index++ {
	}
	t := string(defaultHttpBackendURLBytes[:index])
	// fmt.Println(index, t)
	return t
}

// parseExtraHeaders reads string flags assuming they store extra headers
// with env variable format: --option="KEY0=Value0,Key1=Value1"
func parseExtraHeaders(args []string) map[string][]string {
	out := make(map[string][]string)
	for _, a := range args {
		// create a reader from a string
		reader := strings.NewReader(a)
		m, err := envparse.Parse(reader)
		if err != nil {
			continue
		}
		for k, v := range m {
			out[k] = append(out[k], v)
		}
	}
	return out
}

func init() {
	b := &HttpBackend{}
	RegisterBackend(b)
	SetDefault(b, "enabled", false, "enable the http backend")
	SetDefault(b, "url", defaultHttpBackend.url, "endpoint to send data")
	SetDefault(b, "method", defaultHttpBackend.method, "http method to send data (POST or PUT)")
	SetDefault(b, "header.content-type", defaultHttpBackend.header["Content-Type"], "Content-Type header")
	SetDefault(b, "header.authorization", defaultHttpBackend.header["Authorization"], "Authorization header")
	SetDefault(b, "header.extra", []string{}, "Extra http header with KEY=VALUE format")
}

func (h *HttpBackend) Name() string {
	return "http"
}

func (h *HttpBackend) Init() error {
	logger := GetLogger(h)

	// url
	endpoint, err := GetConfig[string](h, "url")
	if err != nil {
		logger.Warnf("Fail to get http url from config, falling back to '%s'", defaultHttpBackend.url)
		h.url = defaultHttpBackend.url
	} else if _, err := url.ParseRequestURI(endpoint); err != nil {
		logger.Warnf("Input url '%s' is malformed, falling back to '%s'",
			endpoint, defaultHttpBackend.url)
		h.url = defaultHttpBackend.url
	} else {
		h.url = endpoint
	}

	// method
	method, err := GetConfig[string](h, "method")
	if err != nil {
		logger.Warnf("Fail to get http method from config, falling back to '%s'", defaultHttpBackend.method)
		h.method = defaultHttpBackend.method
	} else {
		switch method {
		case POST, PUT:
			h.method = method
		default:
			logger.Warnf("Method '%s' is not supported, falling back to '%s'",
				method, defaultHttpBackend.method)
			h.method = defaultHttpBackend.method
		}
	}

	// headers
	if h.header == nil {
		h.header = make(http.Header)
	}
	// start with extra headers (lower priority)
	extraHeaders, err := GetConfig[[]string](h, "header.extra")
	// fallback to string config
	if len(extraHeaders) == 0 {
		singleExtraHeader, err := GetConfig[string](h, "header.extra")
		if err == nil {
			extraHeaders = []string{singleExtraHeader}
		}
	}
	if err == nil {
		for key, value := range parseExtraHeaders(extraHeaders) {
			canonicalKey := http.CanonicalHeaderKey(key)
			h.header[canonicalKey] = value
		}
	}

	for _, key := range []string{"content-type", "authorization"} {
		canonicalKey := http.CanonicalHeaderKey(key)
		value, err := GetConfig[[]string](h, fmt.Sprintf("header.%s", key))
		if err != nil {
			logger.Warnf("Fail to get http header '%s' from config, falling back to '%v'",
				key, defaultHttpBackend.header[canonicalKey])
			h.header[canonicalKey] = defaultHttpBackend.header[canonicalKey]
		} else {
			h.header[canonicalKey] = value
		}
	}

	return nil
}

func (h *HttpBackend) Close() {

}

func (h *HttpBackend) Write(p *models.Payload) {
	logger := GetLogger(h)

	data, err := json.Marshal(p)
	if err != nil {
		logger.Error(err)
		return
	}

	req, err := http.NewRequest(string(h.method), h.url, bytes.NewReader(data))
	if err != nil {
		logger.Error(err)
		return
	}
	// set headers
	req.Header = h.header

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf("an error occurred while sending data to '%s': %v", h.url, err)
		return
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		logger.Infof("Payload successfully sent to %s (%d)", h.url, resp.StatusCode)
	default:
		logger.Errorf("Unexpected status code: %d", resp.StatusCode)
		length := resp.ContentLength
		if length < 0 {
			length = 512
		}

		buffer := make([]byte, length)
		if n, err := resp.Body.Read(buffer); n > 0 {
			logger.Errorf("response: %s", string(buffer[:n]))
		} else if err != nil {
			logger.Errorf("an error occurred while sending data to '%s': %v", h.url, err)
			return
		}
		// logger.Errorf("error while reaching '%s' (%d): %v", h.url, resp.StatusCode, buffer)
	}

}
