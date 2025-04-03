package backends

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

const ADDR = "127.0.0.1:38081"
const ROUTE = "/api/discovery/situation/"

func postPayload(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Headers: %v\n", r.Header)
	w.WriteHeader(201)
}

func runServer(wg *sync.WaitGroup) *http.Server {
	srv := &http.Server{Addr: ADDR, ReadHeaderTimeout: 3 * time.Second}

	http.HandleFunc(ROUTE, postPayload)

	go func() {
		defer wg.Done() // let main know we are done cleaning up

		// always returns error. ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// unexpected error. port in use?
			fmt.Printf("ListenAndServe(): %v\n", err)
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}
