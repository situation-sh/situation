package store

import (
	"testing"

	"github.com/situation-sh/situation/pkg/test"
)

type GenericStoreTester struct {
	s Store
	t *testing.T
}

func (gst *GenericStoreTester) TestClear() {
	gst.s.Clear()
	// test with IterateMachines
	count := 0
	for range gst.s.IterateMachines() {
		count += 1
	}
	if count != 0 {
		gst.t.Errorf("bad number of machines: %v != 0", count)
	}
}

func (gst *GenericStoreTester) TestGetHost() {
	m := test.RandomHostMachine()
	gst.s.SetHost(m)
	gst.s.InsertMachine(m)
	h := gst.s.GetHost()
	if h == nil {
		gst.t.Errorf("nil host machine")
	}
	if h != m {
		gst.t.Errorf("bad host machine: %v != %v", h, m)
	}
	gst.s.Clear()

	h = gst.s.GetHost()
	if h != nil {
		gst.t.Errorf("bad host machine: %v != nil", h)
	}
}

func (gst *GenericStoreTester) TestGetMachineByNetwork() {
	m0 := test.RandomMachine()
	// try with nil store
	m00 := gst.s.GetMachineByNetwork(m0.NICS[0].IP, m0.NICS[0].MAC)
	if m00 != nil {
		gst.t.Errorf("bad machine: %v != nil", m00)
	}
	// now insert
	gst.s.InsertMachine(m0)

	m1 := gst.s.GetMachineByIP(m0.NICS[0].IP)
	if m1 != m0 {
		gst.t.Errorf("bad machine: %v != %v", m1, m0)
	}

	m2 := gst.s.GetMachineByMAC(m0.NICS[0].MAC)
	if m2 != m0 {
		gst.t.Errorf("bad machine: %v != %v", m2, m0)
	}

	m3 := gst.s.GetMachineByNetwork(m0.NICS[0].IP, m0.NICS[0].MAC)
	if m3 != m0 {
		gst.t.Errorf("bad machine: %v != %v", m3, m0)
	}

	r0 := test.RandomMachine()
	// with low probability this test can fail
	r1 := gst.s.GetMachineByNetwork(r0.NICS[0].IP, r0.NICS[0].MAC)
	if r1 != nil {
		gst.t.Errorf("bad machine: %v != nil (however this test may fail with a very low probability)", r1)
	}

	gst.s.Clear()
}

func (gst *GenericStoreTester) TestGetMachineByHostID() {
	m0 := test.RandomMachine()
	gst.s.InsertMachine(m0)

	m1 := gst.s.GetMachineByHostID(m0.HostID)
	if m1 != m0 {
		gst.t.Errorf("bad machine: %v != %v", m1, m0)
	}

	if m := gst.s.GetMachineByHostID(""); m != nil {
		gst.t.Errorf("bad machine: %v != nil", m)
	}

	if m := gst.s.GetMachineByHostID("???"); m != nil {
		gst.t.Errorf("bad machine: %v != nil", m)
	}
}

func (gst *GenericStoreTester) TestGetMachinesByOpenTCPPort() {
	m0 := test.RandomMachine()
	ports := make([]uint16, 0)
	// ensure all the apps listen on TCP
	for _, app := range m0.Applications() {
		for _, e := range app.Endpoints {
			e.Protocol = "tcp"
			ports = append(ports, e.Port)
		}
	}
	gst.s.InsertMachine(m0)

	for _, p := range ports {
		machines, apps, endpoints := gst.s.GetMachinesByOpenTCPPort(p)
		if len(machines) > 0 && len(apps) > 0 && len(endpoints) > 0 {
			continue
		}
		gst.t.Errorf("no machines/apps seems to match (machines: %+v, apps: %+v",
			machines, apps)
	}

	gst.s.Clear()
}

// func (gst *GenericStoreTester) TestInitPayload() {
// 	n := 10
// 	for i := 0; i < 10; i++ {
// 		gst.s.InsertMachine(test.RandomMachine())
// 	}
// 	p := gst.s.InitPayload()
// 	if p == nil {
// 		gst.t.Errorf("nil payload")
// 	}

// 	//lint:ignore SA5011 no pointer dereference here
// 	m := len(p.Machines)
// 	if m != n {
// 		gst.t.Errorf("bad number of machines: %v != %v", m, n)
// 	}

// 	// print
// 	gst.s.Print()
// 	gst.s.Clear()
// }

func (gst *GenericStoreTester) TestIterateMachines() {
	n := 10
	for i := 0; i < 10; i++ {
		gst.s.InsertMachine(test.RandomMachine())
	}

	count := 0
	for range gst.s.IterateMachines() {
		count += 1
	}

	if count != n {
		gst.t.Errorf("bad number of machines: %v != %v", count, n)
	}

	gst.s.Clear()
}

func (gst *GenericStoreTester) TestGetAllIPv4Networks() {
	n := 4
	for i := 0; i < n; i++ {
		gst.s.InsertMachine(test.RandomMachine())
	}

	networks := gst.s.GetAllIPv4Networks()
	for m := range gst.s.IterateMachines() {
		for _, nic := range m.NICS {
			ok := false
			for _, n := range networks {
				if n.Contains(nic.IP) {
					ok = true
					break
				}
			}
			if !ok {
				gst.t.Errorf("NIC with IP %v does not belong to a network: %v", nic.IP, networks)
			}
		}
	}
}
