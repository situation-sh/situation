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

func testEnableBackend(b Backend) error {
	config.Set(enabledBackendKey(b), "true")
	defer config.Set(enabledBackendKey(b), "false")

	if err := Init(); err != nil {
		return err
	}

	for _, backend := range backends {
		if isEnabled(backend) && backend.Name() == b.Name() {
			return nil
		} else if isEnabled(backend) {
			return fmt.Errorf("backend %s should not be enabled", backend.Name())
		}
	}
	return nil
}

func TestEnableBackend(t *testing.T) {
	for _, backend := range backends {
		if err := testEnableBackend(backend); err != nil {
			t.Errorf("error while enabling backend %s: %v", backend.Name(), err)
		}
	}

}
