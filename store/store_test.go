package store

import (
	"testing"

	"github.com/situation-sh/situation/test"
)

func TestClear(t *testing.T) {
	Clear()
	if store != nil {
		t.Errorf("store is not nil: %v", store)
	}
}

func TestGetHost(t *testing.T) {
	m := test.RandomHostMachine()
	InsertMachine(m)
	h := GetHost()
	if h == nil {
		t.Errorf("nil host machine")
	}
	if h != m {
		t.Errorf("bad host machine: %v != %v", h, m)
	}
	Clear()

	h = GetHost()
	if h != nil {
		t.Errorf("bad host machine: %v != nil", h)
	}
}

func TestGetMachineByNetwork(t *testing.T) {
	m0 := test.RandomMachine()
	// try with nil store
	m00 := GetMachineByNetwork(m0.NICS[0].IP, m0.NICS[0].MAC)
	if m00 != nil {
		t.Errorf("bad machine: %v != nil", m00)
	}
	// now insert
	InsertMachine(m0)

	m1 := GetMachineByIP(m0.NICS[0].IP)
	if m1 != m0 {
		t.Errorf("bad machine: %v != %v", m1, m0)
	}

	m2 := GetMachineByMAC(m0.NICS[0].MAC)
	if m2 != m0 {
		t.Errorf("bad machine: %v != %v", m2, m0)
	}

	m3 := GetMachineByNetwork(m0.NICS[0].IP, m0.NICS[0].MAC)
	if m3 != m0 {
		t.Errorf("bad machine: %v != %v", m3, m0)
	}

	r0 := test.RandomMachine()
	// with low probability this test can fail
	r1 := GetMachineByNetwork(r0.NICS[0].IP, r0.NICS[0].MAC)
	if r1 != nil {
		t.Errorf("bad machine: %v != nil (however this test may fail with a very low probability)", r1)
	}

	Clear()
}

func TestGetMachineByHostID(t *testing.T) {
	m0 := test.RandomMachine()
	InsertMachine(m0)

	m1 := GetMachineByHostID(m0.HostID)
	if m1 != m0 {
		t.Errorf("bad machine: %v != %v", m1, m0)
	}

	if m := GetMachineByHostID(""); m != nil {
		t.Errorf("bad machine: %v != nil", m)
	}

	if m := GetMachineByHostID("???"); m != nil {
		t.Errorf("bad machine: %v != nil", m)
	}

}

func TestGetMachinesByOpenTCPPort(t *testing.T) {
	m0 := test.RandomMachine()
	ports := make([]uint16, 0)
	// ensure all the apps listen on TCP
	for _, app := range m0.Applications() {
		for _, e := range app.Endpoints {
			e.Protocol = "tcp"
			ports = append(ports, e.Port)
		}
	}
	InsertMachine(m0)

	for _, p := range ports {
		machines, apps, endpoints := GetMachinesByOpenTCPPort(p)
		if len(machines) > 0 && len(apps) > 0 && len(endpoints) > 0 {
			continue
		}
		t.Errorf("no machines/apps seems to match (machines: %+v, apps: %+v",
			machines, apps)
	}

	Clear()
}

func TestInitPayload(t *testing.T) {
	n := 10
	for i := 0; i < 10; i++ {
		InsertMachine(test.RandomMachine())
	}
	p := InitPayload()
	if p == nil {
		t.Errorf("nil payload")
	}

	//lint:ignore SA5011 no pointer dereference here
	m := len(p.Machines)
	if m != n {
		t.Errorf("bad number of machines: %v != %v", m, n)
	}

	// print
	Print()
	Clear()
}

func TestIterateMachines(t *testing.T) {
	n := 10
	for i := 0; i < 10; i++ {
		InsertMachine(test.RandomMachine())
	}

	count := 0
	for range IterateMachines() {
		count += 1
	}

	if count != n {
		t.Errorf("bad number of machines: %v != %v", count, n)
	}

	Clear()
}

func TestGetAllIPv4Networks(t *testing.T) {
	n := 4
	for i := 0; i < n; i++ {
		InsertMachine(test.RandomMachine())
	}

	networks := GetAllIPv4Networks()
	for m := range IterateMachines() {
		for _, nic := range m.NICS {
			ok := false
			for _, n := range networks {
				if n.Contains(nic.IP) {
					ok = true
					break
				}
			}
			if !ok {
				t.Errorf("NIC with IP %v does not belong to a network: %v", nic.IP, networks)
			}
		}

	}

}
