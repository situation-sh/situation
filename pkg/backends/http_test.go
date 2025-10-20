package backends

// import (
// 	"fmt"
// 	"net/http"
// 	"time"

// 	"github.com/situation-sh/situation/pkg/config"
// )

// const ADDR = "127.0.0.1:38081"
// const ROUTE = "/api/discovery/situation/"

// const urlKey = "backends.http.url"

// type httpBackendTestServer struct {
// 	initialURL string
// 	srv        *http.Server
// 	log        func(msg string, args ...any)
// }

// func (s *httpBackendTestServer) handlerFactory() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		s.log("Headers: %v\n", r.Header)
// 		w.WriteHeader(201)
// 	}
// }

// func (s *httpBackendTestServer) stop() error {
// 	if s.initialURL != "" {
// 		config.Set(urlKey, s.initialURL)
// 	}
// 	if s.srv != nil {
// 		if err := s.srv.Shutdown(nil); err != nil {
// 			return fmt.Errorf("error while shutting down server: %w", err)
// 		}
// 	}
// 	return nil
// }

// func (s *httpBackendTestServer) setUp() error {
// 	initialURL, err := config.Get[string](urlKey)
// 	if err != nil {
// 		return err
// 	}
// 	s.initialURL = initialURL
// 	if err := config.Set(urlKey, "http://"+ADDR+ROUTE); err != nil {
// 		return err
// 	}
// 	if err := config.Set("backends.http.extra-header", "X-Organization-ID=situation.sh"); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (s *httpBackendTestServer) tearDown() error {
// 	if err := config.Set(urlKey, s.initialURL); err != nil {
// 		return err
// 	}
// 	if err := config.Set("backends.http.extra-header", ""); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (s *httpBackendTestServer) start() error {
// 	if err := s.setUp(); err != nil {
// 		return fmt.Errorf("error while setting up server: %w", err)
// 	}

// 	s.srv = &http.Server{Addr: ADDR, ReadHeaderTimeout: 3 * time.Second}
// 	mux := http.NewServeMux()
// 	mux.HandleFunc(ROUTE, s.handlerFactory())
// 	s.srv.Handler = mux

// 	go func() {
// 		// always returns error. ErrServerClosed on graceful close
// 		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
// 			// unexpected error. port in use?
// 			fmt.Printf("ListenAndServe(): %v\n", err)
// 		}

// 		if err := s.tearDown(); err != nil {
// 			fmt.Printf("error while tearing down server: %v\n", err)
// 		}

// 	}()

// 	// returning reference so caller can call Shutdown()
// 	return nil
// }
