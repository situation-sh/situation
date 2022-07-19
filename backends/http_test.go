package backends

import (
	"fmt"
	"net/http"
	"sync"
)

func postPayload(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(201)
}

func runServer(wg *sync.WaitGroup) *http.Server {
	srv := &http.Server{Addr: "127.0.0.1:38080"}

	http.HandleFunc("/api/discovery/situation/", postPayload)

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
