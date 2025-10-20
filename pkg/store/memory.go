package store

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/utils"
)

type MemoryStore struct {
	mu    sync.Mutex
	agent uuid.UUID
	store []*models.Machine
}

func NewMemoryStore(agent uuid.UUID) *MemoryStore {
	return &MemoryStore{
		store: make([]*models.Machine, 0),
		agent: agent,
	}
}

func (ds *MemoryStore) Open() error {
	return nil
}

func (ds *MemoryStore) Close() error {
	return nil
}

func (ds *MemoryStore) Clear() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.store = nil
}

func (ds *MemoryStore) GetHost() *models.Machine {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	for _, m := range ds.store {
		if ds.IsHost(m) {
			return m
		}
	}
	return nil
}

func (ds *MemoryStore) SetHost(m *models.Machine) {
	if m == nil {
		return
	}
	// set the agent ID
	var u uuid.UUID
	copy(u[:], ds.agent[:])
	m.Agent = &u
}

func (ds *MemoryStore) IsHost(m *models.Machine) bool {
	if m == nil {
		return false
	}
	if m.Agent == nil {
		return false
	}
	return m.Agent.String() == ds.agent.String()
}

func (ds *MemoryStore) GetMachineByNetwork(ip net.IP, mac net.HardwareAddr) *models.Machine {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if len(ds.store) == 0 {
		return nil
	}
	for _, m := range ds.store {
		for _, nic := range m.NICS {
			if nic.Match(ip, mac) {
				return m
			}
		}
	}
	return nil
}

func (ds *MemoryStore) GetMachineByMAC(mac net.HardwareAddr) *models.Machine {
	return ds.GetMachineByNetwork(nil, mac)
}

func (ds *MemoryStore) GetMachineByIP(ip net.IP) *models.Machine {
	return ds.GetMachineByNetwork(ip, nil)
}

func (ds *MemoryStore) GetMachineByHostID(id string) *models.Machine {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if id == "" {
		return nil
	}
	for _, m := range ds.store {
		if m.HostID == id {
			return m
		}
	}
	return nil
}

func (ds *MemoryStore) GetMachinesByOpenTCPPort(port uint16) ([]*models.Machine, []*models.Application, []*models.ApplicationEndpoint) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	outMachines := make([]*models.Machine, 0)
	outApps := make([]*models.Application, 0)
	outEndpoints := make([]*models.ApplicationEndpoint, 0)
	for _, machine := range ds.store {
		for _, app := range machine.Applications() {
			for _, endpoint := range app.Endpoints {
				if endpoint.Protocol == "tcp" && endpoint.Port == port {
					outMachines = append(outMachines, machine)
					outApps = append(outApps, app)
					outEndpoints = append(outEndpoints, endpoint)
				}
			}
		}
	}
	return outMachines, outApps, outEndpoints
}

func (ds *MemoryStore) InitPayload() *models.Payload {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	return &models.Payload{
		Machines: ds.store,
	}
}

func (ds *MemoryStore) InsertMachine(m *models.Machine) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.store = append(ds.store, m)
}

func (ds *MemoryStore) Print() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	bytes, _ := json.Marshal(ds.store)
	fmt.Println(string(bytes))
}

func (ds *MemoryStore) IterateMachines() chan *models.Machine {
	c := make(chan *models.Machine)
	go func() {
		ds.mu.Lock()
		defer ds.mu.Unlock()
		for _, m := range ds.store {
			c <- m
		}
		close(c)
	}()
	return c
}

func (ds *MemoryStore) GetAllIPv4Networks() []*net.IPNet {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	mapper := make(map[string]*net.IPNet)
	out := make([]*net.IPNet, 0)
	for _, m := range ds.store {
		for _, nic := range m.NICS {
			if nic.IP != nil && nic.MaskSize > 0 {
				nw := nic.Network()
				network := utils.EnforceMask(nw)
				cidr := network.String()
				if _, exists := mapper[cidr]; !exists {
					mapper[cidr] = network
				}
			}
		}
	}
	for _, n := range mapper {
		out = append(out, n)
	}
	return out
}

// Clear does the job
// func Clear() {
// 	internalMutex.Lock()
// 	defer internalMutex.Unlock()

// 	// see: https://stackoverflow.com/a/16973160
// 	// Setting the slice to nil is the best way to clear a slice.
// 	// nil slices in go are perfectly well behaved and setting the
// 	// slice to nil will release the underlying memory to the garbage
// 	// collector.
// 	store = nil
// }

// func GetHost() *models.Machine {
// 	internalMutex.Lock()
// 	defer internalMutex.Unlock()

// 	for _, m := range store {
// 		if m.IsHost() {
// 			return m
// 		}
// 	}
// 	return nil
// }

// // GetMachineByNetwork returns the first machine with network attributes
// // that match the input (if non-nil). nil-input are ignored
// func GetMachineByNetwork(ip net.IP, mac net.HardwareAddr) *models.Machine {
// 	internalMutex.Lock()
// 	defer internalMutex.Unlock()

// 	if len(store) == 0 {
// 		return nil
// 	}

// 	for _, m := range store {
// 		for _, nic := range m.NICS {
// 			// check network
// 			if nic.Match(ip, mac) {
// 				return m
// 			}
// 		}
// 	}
// 	return nil
// }

// func GetMachineByMAC(mac net.HardwareAddr) *models.Machine {
// 	return GetMachineByNetwork(nil, mac)
// }

// func GetMachineByIP(ip net.IP) *models.Machine {
// 	return GetMachineByNetwork(ip, nil)
// }

// func GetMachineByHostID(id string) *models.Machine {
// 	internalMutex.Lock()
// 	defer internalMutex.Unlock()
// 	// ignore empty ID
// 	if id == "" {
// 		return nil
// 	}

// 	for _, m := range store {
// 		if m.HostID == id {
// 			return m
// 		}
// 	}

// 	return nil
// }

// // GetMachinesByOpenTCPPort returns the list of machines that have
// // an application listening on this TCP port. In addition it also
// // returns the list of the related app endpoints.
// func GetMachinesByOpenTCPPort(port uint16) ([]*models.Machine, []*models.Application, []*models.ApplicationEndpoint) {
// 	internalMutex.Lock()
// 	defer internalMutex.Unlock()

// 	outMachines := make([]*models.Machine, 0)
// 	outApps := make([]*models.Application, 0)
// 	outEndpoints := make([]*models.ApplicationEndpoint, 0)

// 	for _, machine := range store {
// 		for _, app := range machine.Applications() {
// 			for _, endpoint := range app.Endpoints {
// 				if endpoint.Protocol == "tcp" && endpoint.Port == port {
// 					outMachines = append(outMachines, machine)
// 					outApps = append(outApps, app)
// 					outEndpoints = append(outEndpoints, endpoint)
// 				}
// 			}
// 		}
// 	}
// 	return outMachines, outApps, outEndpoints
// }

// func InitPayload() *models.Payload {
// 	return &models.Payload{
// 		Machines: store,
// 	}
// }

// func InsertMachine(m *models.Machine) {
// 	internalMutex.Lock()
// 	defer internalMutex.Unlock()
// 	store = append(store, m)
// }

// func Print() {
// 	bytes, _ := json.Marshal(store)
// 	fmt.Println(string(bytes))
// }

// func IterateMachines() chan *models.Machine {
// 	c := make(chan *models.Machine)

// 	go func() {
// 		for _, m := range store {
// 			c <- m
// 		}
// 		close(c)
// 	}()

// 	return c
// }

// func GetAllIPv4Networks() []*net.IPNet {
// 	mapper := make(map[string]*net.IPNet)
// 	out := make([]*net.IPNet, 0)

// 	for _, m := range store {
// 		for _, nic := range m.NICS {
// 			if nic.IP != nil && nic.MaskSize > 0 {
// 				nw := nic.Network()
// 				network := utils.EnforceMask(nw)
// 				cidr := network.String()
// 				if _, exists := mapper[cidr]; !exists {
// 					mapper[cidr] = network
// 				}
// 			}
// 		}
// 	}

// 	// return list of networks
// 	for _, n := range mapper {
// 		out = append(out, n)
// 	}

// 	return out
// }
