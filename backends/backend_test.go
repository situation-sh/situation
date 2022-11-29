package backends

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/situation-sh/situation/models"
)

// GenericTestBackend is a basic function to test a backend
func GenericTestBackend(b Backend, payload *models.Payload) error {
	// fake an http server
	if b.Name() == "http" {
		defaultHttpBackend.url = "http://127.0.0.1:38080/api/discovery/situation/"
		wg := sync.WaitGroup{}
		wg.Add(1)
		srv := runServer(&wg)
		defer func() {
			if err := srv.Shutdown(context.TODO()); err != nil {
				fmt.Println(err)
			}
			wg.Wait()
		}()
	}
	if err := b.Init(); err != nil {
		return err
	}

	b.Write(payload)
	b.Close()
	return nil
}

func TestBackends(t *testing.T) {
	p := models.Payload{}
	if err := gofakeit.Struct(&p); err != nil {
		t.Error(err)
	}
	// p := test.RandomPayload()
	// p := randomPayload()
	for name, b := range backends {
		fmt.Printf("--- BACKEND: %s\n", name)
		if err := GenericTestBackend(b, &p); err != nil {
			t.Errorf("error with backend %s: %v", name, err)
		}
	}
}
