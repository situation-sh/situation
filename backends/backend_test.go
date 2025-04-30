package backends

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/test"
)

// GenericTestBackend is a basic function to test a backend
func GenericTestBackend(b Backend, payload *models.Payload) error {

	// fake an http server
	// if b.Name() == "http" {
	// 	fs := flag.NewFlagSet("test", flag.ExitOnError)
	// 	fs.Bool("backends.http.enabled", true, "")
	// 	fs.String("backends.http.header.extra", "X-WTF-ID=89", "")
	// 	// fs.("backends.http.header.extra", []string{"X-WTF-ID=89"}, "")
	// 	ctx := cli.NewContext(nil, fs, nil)

	// 	config.InjectContext(ctx)

	// 	if !isEnabled(b) {
	// 		return fmt.Errorf("backend %s is not enabled", b.Name())
	// 	}

	// 	defaultHttpBackend.url = fmt.Sprintf("http://%s%s", ADDR, ROUTE)
	// 	wg := sync.WaitGroup{}
	// 	wg.Add(1)
	// 	srv := runServer(&wg)
	// 	defer func() {
	// 		if err := srv.Shutdown(context.Background()); err != nil {
	// 			fmt.Println(err)
	// 		}
	// 		wg.Wait()
	// 	}()
	// }
	if err := b.Init(); err != nil {
		return err
	}

	b.Write(payload)
	b.Close()
	return nil
}

func TestBackends(t *testing.T) {
	p := test.RandomPayload()
	for name, b := range backends {
		fmt.Printf("--- BACKEND: %s\n", name)
		if err := GenericTestBackend(b, p); err != nil {
			t.Errorf("error with backend %s: %v", name, err)
		}
	}
}

// func TestPrepare(t *testing.T) {
// 	proxy := &StdoutBackend{}
// 	name := proxy.Name()
// 	backend, exists := backends[name]
// 	if !exists {
// 		t.Errorf("the backend %s is not registered", name)
// 	}
// 	key := fmt.Sprintf("backends.%s.enabled", name)

// 	fs := flag.NewFlagSet("test", flag.ExitOnError)
// 	// define a flag equal to false by default
// 	fs.Bool(key, false, "")
// 	// fake the cmdline option
// 	fs.Set(key, "true")

// 	c := cli.NewContext(nil, fs, nil)
// 	if !c.IsSet(key) {
// 		t.Errorf("the key %s not set in context", key)
// 	}

// 	config.InjectContext(c)

// 	if !isEnabled(backend) {
// 		t.Errorf("backend %s is not enabled", name)
// 	}

// 	prepareBackends()
// 	if len(enabledBackends) != 1 {
// 		t.Errorf("1 backend expected, got %v", enabledBackends)
// 	}

// 	if err := Init(); err != nil {
// 		t.Error(err)
// 	}
// 	Write(&models.Payload{})
// 	Close()
// }

func TestNetworkInterfaceUnmarshal(t *testing.T) {
	nic := test.RandomNIC()
	raw, err := json.Marshal(nic)
	if err != nil {
		t.Fatalf("error while marshalling NIC: %v", err)
	}

	var otherNic models.NetworkInterface
	err = json.Unmarshal(raw, &otherNic)
	if err != nil {
		t.Fatalf("error while unmarshalling NIC: %v", err)
	}

	if nic.Name != otherNic.Name {
		t.Errorf("bad name %s != %s", nic.Name, otherNic.Name)
	}
	if nic.MAC.String() != otherNic.MAC.String() {
		t.Errorf("bad mac %v != %v", nic.MAC, otherNic.MAC)
	}
	if !nic.IP.Equal(otherNic.IP) {
		t.Errorf("bad ip %v != %v", nic.IP, otherNic.IP)
	}
	if !nic.IP6.Equal(otherNic.IP6) {
		t.Errorf("bad ip6 %v != %v", nic.IP6, otherNic.IP6)
	}
	if !nic.Gateway.Equal(otherNic.Gateway) {
		t.Errorf("bad gateway %v != %v", nic.Gateway, otherNic.Gateway)
	}
	if *nic.Flags != *otherNic.Flags {
		t.Errorf("bad flags %+v != %+v", nic.Flags, otherNic.Flags)
	}

	t.Logf("\n%+v\n", otherNic)
}
