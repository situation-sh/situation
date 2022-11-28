package store

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/situation-sh/situation/models"
)

var store []*models.Machine

// internalMutex is used to make the store
// thread safe
// All the getters and the setters MUST use it
var internalMutex sync.Mutex

func init() {
	store = make([]*models.Machine, 0)
}

// Clear does the job
func Clear() {
	internalMutex.Lock()
	defer internalMutex.Unlock()

	// see: https://stackoverflow.com/a/16973160
	// Setting the slice to nil is the best way to clear a slice.
	// nil slices in go are perfectly well behaved and setting the
	// slice to nil will release the underlying memory to the garbage
	// collector.
	store = nil
}

func GetHost() *models.Machine {
	internalMutex.Lock()
	defer internalMutex.Unlock()

	for _, m := range store {
		if m.IsHost() {
			return m
		}
	}
	return nil
}

// GetMachineByNetwork returns the first machine with network attributes
// that match the input (if non-nil). nil-input are ignored
func GetMachineByNetwork(ip net.IP, mac net.HardwareAddr) *models.Machine {
	internalMutex.Lock()
	defer internalMutex.Unlock()

	if len(store) == 0 {
		return nil
	}

	for _, m := range store {
		for _, nic := range m.NICS {
			// check network
			if nic.Match(ip, mac) {
				return m
			}
		}
	}
	return nil
}

func GetMachineByMAC(mac net.HardwareAddr) *models.Machine {
	return GetMachineByNetwork(nil, mac)
}

func GetMachineByIP(ip net.IP) *models.Machine {
	return GetMachineByNetwork(ip, nil)
}

func GetMachineByHostID(id string) *models.Machine {
	// ignore empty ID
	if id == "" {
		return nil
	}

	for _, m := range store {
		if m.HostID == id {
			return m
		}
	}

	return nil
}

// GetMachinesByOpenTCPPort returns the list of machines that have
// an application listening on this TCP port. In addition it also
// returns the list of the related app endpoints.
func GetMachinesByOpenTCPPort(port uint16) ([]*models.Machine, []*models.ApplicationEndpoint) {
	outMachines := make([]*models.Machine, 0)
	outEndpoints := make([]*models.ApplicationEndpoint, 0)

	for _, machine := range store {
		for _, app := range machine.Applications() {
			for _, endpoint := range app.Endpoints {
				if endpoint.Protocol == "tcp" && endpoint.Port == port {
					outMachines = append(outMachines, machine)
					outEndpoints = append(outEndpoints, endpoint)
				}
			}
		}
	}
	return outMachines, outEndpoints
}

func InitPayload() *models.Payload {
	return &models.Payload{
		Machines: store,
	}
}

func InsertMachine(m *models.Machine) {
	internalMutex.Lock()
	defer internalMutex.Unlock()
	store = append(store, m)
}

func Print() {
	bytes, _ := json.Marshal(store)
	fmt.Println(string(bytes))
}

func IterateMachines() chan *models.Machine {
	c := make(chan *models.Machine)

	go func() {
		for _, m := range store {
			c <- m
		}
		close(c)
	}()

	return c
}
