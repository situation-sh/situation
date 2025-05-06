package backends

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/situation-sh/situation/config"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/test"
)

// GenericTestBackend is a basic function to test a backend
func GenericTestBackend(b Backend, payload *models.Payload) error {
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

func TestPackage(t *testing.T) {
	// activate all the backends
	for _, backend := range backends {
		config.Set(enableBackendKey(backend), "true")
		defer config.Set(enableBackendKey(backend), "false")
	}

	// save file to tmp
	initialPath, err := config.Get[string]("backends.file.path")
	if err != nil {
		t.Fatalf("error while getting backends.file.path: %v", err)
	}
	config.Set("backends.file.path", t.TempDir()+"/situation.json")
	defer config.Set("backends.file.path", initialPath)

	// start dummy http server (it also configure the backend)
	srv := &httpBackendTestServer{log: t.Logf}
	if err := srv.start(); err != nil {
		t.Fatalf("error while starting HTTP backend server: %v", err)
	}
	defer srv.stop()

	if err := Init(); err != nil {
		t.Error(err)
	}

	r := test.RandomPayload()
	if err := Write(r); err != nil {
		t.Error(err)
	}

	if err := Close(); err != nil {
		t.Error(err)
	}
}

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
