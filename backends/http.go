package backends

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/situation-sh/situation/config"
	"github.com/situation-sh/situation/models"
)

type HttpBackend struct {
	URL           string
	Method        string
	ContentType   []string
	Authorization []string
	ExtraHeaders  []string
}

func (h *HttpBackend) populateHeaders(headers *http.Header) error {
	for _, ct := range h.ContentType {
		headers.Add("content-type", ct)
	}
	for _, auth := range h.Authorization {
		headers.Add("authorization", auth)
	}
	for _, extra := range h.ExtraHeaders {
		s := strings.Split(extra, "=")
		if len(s) != 2 {
			return fmt.Errorf("invalid extra header format: %s", extra)
		}
		headers.Add(s[0], s[1])
	}
	headers.Add("User-Agent", fmt.Sprintf("situation/%s", config.Version))
	return nil
}

func init() {
	b := &HttpBackend{
		URL:           "http://127.0.0.1:8000/import/situation/",
		Method:        "POST",
		ContentType:   []string{"application/json"},
		Authorization: []string{config.GetAgent().String()},
		ExtraHeaders:  []string{},
	}
	RegisterBackend(b)
	SetDefault(b, "url", &b.URL, "endpoint to send data")
	SetDefault(b, "method", &b.Method, "http method to send data (POST or PUT)")
	SetDefault(b, "content-type", &b.ContentType, "Content-Type header")
	SetDefault(b, "authorization", &b.Authorization, "Authorization header")
	SetDefault(b, "extra-header", &b.ExtraHeaders, "Extra http header with KEY=VALUE format")
}

func (h *HttpBackend) Name() string {
	return "http"
}

func (h *HttpBackend) Init() error {
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

	req, err := http.NewRequest(h.Method, h.URL, bytes.NewReader(data))
	if err != nil {
		logger.Error(err)
		return
	}

	// provide headers based on config
	if err := h.populateHeaders(&req.Header); err != nil {
		logger.Error(err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf("an error occurred while sending data to '%s': %v", h.URL, err)
		return
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		logger.Infof("Payload successfully sent to %s (%d)", h.URL, resp.StatusCode)
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
			logger.Errorf("an error occurred while sending data to '%s': %v", h.URL, err)
			return
		}

	}

}
