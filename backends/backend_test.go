package backends

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/situation-sh/situation/config"
	"github.com/situation-sh/situation/models"
	"github.com/urfave/cli/v2"
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

func TestPrepare(t *testing.T) {
	proxy := &StdoutBackend{}
	name := proxy.Name()
	backend, exists := backends[name]
	if !exists {
		t.Errorf("the backend %s is not registered", name)
	}
	key := fmt.Sprintf("backends.%s.enabled", name)

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	// define a flag equal to false by default
	fs.Bool(key, false, "")
	// fake the cmdline option
	fs.Set(key, "true")

	c := cli.NewContext(nil, fs, nil)
	if !c.IsSet(key) {
		t.Errorf("the key %s not set in context", key)
	}

	config.InjectContext(c)

	if !isEnabled(backend) {
		t.Errorf("backend %s is not enabled", name)
	}

	prepareBackends()
	if len(enabledBackends) != 1 {
		t.Errorf("1 backend expected, got %v", enabledBackends)
	}

	if err := Init(); err != nil {
		t.Error(err)
	}
	Write(&models.Payload{})
	Close()
}
