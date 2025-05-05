package backends

import (
	"fmt"
	"net/http"
	"time"

	"github.com/situation-sh/situation/config"
)

const ADDR = "127.0.0.1:38081"
const ROUTE = "/api/discovery/situation/"

const urlKey = "backends.http.url"

type httpBackendTestServer struct {
	initialURL string
	srv        *http.Server
}

func postPayload(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Headers: %v\n", r.Header)
	w.WriteHeader(201)
}

func (s *httpBackendTestServer) stop() error {
	if s.initialURL != "" {
		config.Set(urlKey, s.initialURL)
	}
	if s.srv != nil {
		if err := s.srv.Shutdown(nil); err != nil {
			return fmt.Errorf("error while shutting down server: %w", err)
		}
	}
	return nil
}

func (s *httpBackendTestServer) start() error {
	initialURL, err := config.Get[string](urlKey)
	if err != nil {
		return err
	}
	s.initialURL = initialURL
	config.Set(urlKey, "http://"+ADDR+ROUTE)

	s.srv = &http.Server{Addr: ADDR, ReadHeaderTimeout: 3 * time.Second}
	mux := http.NewServeMux()
	mux.HandleFunc(ROUTE, postPayload)
	s.srv.Handler = mux

	go func() {
		defer config.Set(urlKey, initialURL) // restore original URL

		// always returns error. ErrServerClosed on graceful close
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			// unexpected error. port in use?
			fmt.Printf("ListenAndServe(): %v\n", err)
		}
	}()

	// returning reference so caller can call Shutdown()
	return nil
}
