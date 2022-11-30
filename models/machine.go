package models

import (
	"math/rand"
	"net"
	"time"

	"github.com/google/uuid"
)

// Machine is a generic structure to represent nodes on
// an information system. It can be a physical machine,
// a VM, a container...
type Machine struct {
	InternalID          int                 `json:"internal_id"`
	Hostname            string              `json:"hostname,omitempty"`
	HostID              string              `json:"host_id,omitempty"`
	Arch                string              `json:"arch,omitempty"`
	Platform            string              `json:"platform,omitempty"`
	Distribution        string              `json:"distribution,omitempty"`
	DistributionVersion string              `json:"distribution_version,omitempty"`
	ParentMachine       int                 `json:"parent_machine,omitempty"`
	CPU                 *CPU                `json:"cpu,omitempty"`
	NICS                []*NetworkInterface `json:"nics"`
	Packages            []*Package          `json:"packages"`
	Disks               []*Disk             `json:"disks"`
	GPUS                []*GPU              `json:"gpus"`
	Agent               *uuid.UUID          `json:"hosted_agent,omitempty"`
	Uptime              time.Duration       `json:"uptime,omitempty"`
}

// NewMachine inits a new Machine structure
func NewMachine() *Machine {
	return &Machine{
		InternalID: 1 + rand.Intn(65535), // ensure the InternalId is greater than 1
		Packages:   make([]*Package, 0),
		NICS:       make([]*NetworkInterface, 0),
		Disks:      make([]*Disk, 0),
		GPUS:       make([]*GPU, 0),
	}
}

// IsHost returns whether the machine is the
// current machine where this agent runs
func (m *Machine) IsHost() bool {
	return m.Agent != nil
}

// GetNetworkInterfaceByIP returns the network interface of
// this machine that has this IP. It returns nil if the IP
// is nil and if the machine has not this IP
func (m *Machine) GetNetworkInterfaceByIP(ip net.IP) *NetworkInterface {
	if ip == nil {
		return nil
	}
	for _, nic := range m.NICS {
		if ip.Equal(nic.IP) || ip.Equal(nic.IP6) {
			return nic
		}
	}
	return nil
}

// GetNetworkInterfaceByMAC does the same thing as GetNetworkInterfaceByIP
// but with MAC address. It returns nil when IP is nil or when no network interface
// with this MAC has been found
func (m *Machine) GetNetworkInterfaceByMAC(mac net.HardwareAddr) *NetworkInterface {
	if mac == nil {
		return nil
	}
	for _, nic := range m.NICS {
		if mac.String() == nic.MAC.String() {
			return nic
		}
	}
	return nil
}

// HasIP returns whether the machine has a network interface
// with this IP
func (m *Machine) HasIP(ip net.IP) bool {
	return m.GetNetworkInterfaceByIP(ip) != nil
}

// func (m *Machine) lastApplication() *Application {
// 	if len(m.Applications) == 0 {
// 		return nil
// 	}
// 	return m.Applications[len(m.Applications)-1]
// }

// GetOrCreateApplicationByName returns the app running on this machine
// given its name. It creates it if it does exists (a boolean is returned
// to tells if the app has been created or not)
func (m *Machine) GetOrCreateApplicationByName(name string) (*Application, bool) {
	for _, p := range m.Packages {
		for _, s := range p.Applications {
			if s.Name == name {
				return s, false
			}
		}
	}
	app := Application{Name: name}
	pkg := Package{Applications: []*Application{&app}}
	m.Packages = append(m.Packages, &pkg)
	// m.Applications = append(m.Applications, &Application{Name: name})
	return &app, true
}

// GetOrCreateApplicationByEndpoint returns the app running on this machine
// given its port, protocol and address. It creates it if it does exists
// (a boolean is returned to tells if the app has been created or not)
func (m *Machine) GetOrCreateApplicationByEndpoint(port uint16, protocol string, addr net.IP) (*Application, bool) {
	for _, p := range m.Packages {
		for _, s := range p.Applications {
			for _, e := range s.Endpoints {
				if e.Port == port && e.Protocol == protocol && e.Addr.Equal(addr) {
					return s, false
				}
			}
		}
	}
	// create the endpoint
	endpoint := ApplicationEndpoint{
		Port:     port,
		Protocol: protocol,
		Addr:     addr,
	}
	app := Application{Endpoints: []*ApplicationEndpoint{&endpoint}}
	pkg := Package{Applications: []*Application{&app}}
	m.Packages = append(m.Packages, &pkg)
	// m.Applications = append(m.Applications,
	// 	&Application{Endpoints: []*ApplicationEndpoint{&endpoint}})
	return &app, true
}

func (m *Machine) Applications() []*Application {
	out := make([]*Application, 0)
	for _, p := range m.Packages {
		out = append(out, p.Applications...)
	}
	return out
}

func (m *Machine) GetPackageByApplicationPath(path string) *Package {
	for _, p := range m.Packages {
		for _, a := range p.Applications {
			if a.Name == path {
				return p
			}
		}
	}
	return nil
}

// InsertPackage add the given package into the machine.
// It tries to merge with previously created package
// base on application path
// It returns whether the package has been merged
// (otherwise it means that it already exists or
// should not be created)
func (m *Machine) InsertPackage(pkg *Package) (*Package, bool) {
	for _, p := range m.Packages {
		if p.Equal(pkg) {
			// we already have the package
			return p, false
		}
		for _, app := range p.Applications {
			for _, f := range pkg.Files {
				// matching based on application path
				if app.Name == f {
					p.Name = pkg.Name
					p.Version = pkg.Version
					p.Vendor = pkg.Vendor
					p.Manager = pkg.Manager
					copy(p.Files, pkg.Files)
					return p, true
				}
			}
		}
	}
	// otherwise we append the package to the machine
	// or we can do nothing (not to pollute)
	// m.Packages = append(m.Packages, pkg)
	return nil, false
}
