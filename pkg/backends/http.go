package backends

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/asiffer/puzzle"
	"github.com/situation-sh/situation/pkg/models"
)

type HttpBackend struct {
	BaseBackend

	URL           string
	Method        string
	ContentType   []string
	Authorization []string
	ExtraHeaders  []string
}

func (h *HttpBackend) populateHeaders(headers *http.Header) error {
	// headers.Set("user-agent", fmt.Sprintf("situation/%s", Version))
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
	// headers.Add("User-Agent", fmt.Sprintf("situation/%s", config.Version))
	return nil
}

func init() {
	b := &HttpBackend{
		URL:         "http://127.0.0.1:8000/import/situation/",
		Method:      "POST",
		ContentType: []string{"application/json"},
		// Authorization: []string{config.GetAgent().String()},
		ExtraHeaders: []string{},
	}
	registerBackend(b)
	// SetDefault(b, "url", &b.URL, "endpoint to send data")
	// SetDefault(b, "method", &b.Method, "http method to send data (POST or PUT)")
	// SetDefault(b, "content-type", &b.ContentType, "Content-Type header")
	// SetDefault(b, "authorization", &b.Authorization, "Authorization header")
	// SetDefault(b, "extra-header", &b.ExtraHeaders, "Extra http header with KEY=VALUE format")
}

func (h *HttpBackend) Name() string {
	return "http"
}

func (h *HttpBackend) Bind(config *puzzle.Config) error {
	if err := setDefault(config, h, "url", &h.URL, "endpoint to send data"); err != nil {
		return err
	}
	if err := setDefault(config, h, "method", &h.Method, "http method to send data (POST or PUT)"); err != nil {
		return err
	}
	if err := setDefault(config, h, "content-type", &h.ContentType, "Content-Type header"); err != nil {
		return err
	}
	if err := setDefault(config, h, "authorization", &h.Authorization, "Authorization header"); err != nil {
		return err
	}
	if err := setDefault(config, h, "extra-header", &h.ExtraHeaders, "Extra http header with KEY=VALUE format"); err != nil {
		return err
	}
	return nil
}

func (h *HttpBackend) Init() error {
	return nil
}

func (h *HttpBackend) Close() error {
	return nil
}

func (h *HttpBackend) Write(p *models.Payload) error {
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("error while marshalling payload to json: %w", err)
	}

	req, err := http.NewRequest(h.Method, h.URL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error while creating request: %w", err)
	}

	// provide headers based on config
	if err := h.populateHeaders(&req.Header); err != nil {
		return fmt.Errorf("error while populating request headers: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error while sending data to %s: %w", h.URL, err)
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		h.logger.Infof("Payload successfully sent to %s (%d)", h.URL, resp.StatusCode)
	default:
		h.logger.Errorf("Unexpected status code: %d", resp.StatusCode)
		length := resp.ContentLength
		if length < 0 {
			length = 512
		}

		buffer := make([]byte, length)
		if n, err := resp.Body.Read(buffer); n > 0 {
			return fmt.Errorf("response error (%v): %s", resp.StatusCode, string(buffer[:n]))
		} else if err != nil {
			return fmt.Errorf("error while reading body data after failed request: %w", err)
		}

	}
	return nil
}
