package models

import (
	"net"
	"time"

	"github.com/google/uuid"
)

var machineCounter = 0

// Machine is a generic structure to represent nodes on
// an information system. It can be a physical machine,
// a VM, a container...
type Machine struct {
	InternalID          int                 `json:"internal_id" jsonschema:"description=internal reference of the machine (if we want to point to this machine within the json),example=53127,minimum=1"`
	Hostname            string              `json:"hostname,omitempty" jsonschema:"description=name of the machine,example=DESKTOP-2HHPC7I,example=PC-JEAN-LUC,example=server07"`
	HostID              string              `json:"host_id,omitempty" jsonschema:"description=machine uuid identifier,example=8375c6c3-de33-41a4-bdb2-4e467d9f632c"`
	Arch                string              `json:"arch,omitempty" jsonschema:"description=architecture,example=x86_64"`
	Platform            string              `json:"platform,omitempty" jsonschema:"description=system base platform,example=linux,example=windows,example=docker"`
	Distribution        string              `json:"distribution,omitempty" jsonschema:"description=OS name (or base image),example=fedora,example=Microsoft Windows 10 Home,example=postgres"`
	DistributionVersion string              `json:"distribution_version,omitempty" jsonschema:"description=OS version (or image version),example=36,example=10.0.19044 Build 19044,example=latest"`
	Uptime              time.Duration       `json:"uptime,omitempty" jsonschema:"description=machine uptime in nanoseconds,example=22343000000000,example=13521178203519"`
	ParentMachine       int                 `json:"parent_machine,omitempty" jsonschema:"description=internal reference of the parent machine (docker or VM cases especially),example=53127"`
	Agent               *uuid.UUID          `json:"hosted_agent,omitempty" jsonschema:"description=collector identifier (agent case)"`
	CPU                 *CPU                `json:"cpu,omitempty" jsonschema:"description=CPU information"`
	NICS                []*NetworkInterface `json:"nics" jsonschema:"description=list of network devices"`
	Packages            []*Package          `json:"packages" jsonschema:"description=list of packages"`
	Disks               []*Disk             `json:"disks" jsonschema:"description=list of disks"`
	GPUS                []*GPU              `json:"gpus" jsonschema:"description=list of GPU"`
	CPE                 string              `json:"cpe,omitempty" jsonschema:"description=OS CPE uri,example=cpe:2.3:o:microsoft:windows_server_2022:-:*:*:*:datacenter:*:x64:*"`
	Chassis             string              `json:"chassis,omitempty" jsonschema:"description=machine kind,example=vm,example=laptop"`
}

// NewMachine inits a new Machine structure
func NewMachine() *Machine {
	// increment ID
	machineCounter++
	// return new machine
	return &Machine{
		InternalID: machineCounter, // ensure the InternalId is greater than 1
		Packages:   make([]*Package, 0),
		NICS:       make([]*NetworkInterface, 0),
		Disks:      make([]*Disk, 0),
		GPUS:       make([]*GPU, 0),
	}
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
// to tells if the app has been created or not). It also looks at the package
// files if the app has not been found.
func (m *Machine) GetOrCreateApplicationByName(name string) (*Application, bool) {
	var pkg *Package = nil

	for _, p := range m.Packages {
		for _, s := range p.Applications {
			if s.Name == name {
				return s, false
			}
		}
	}

	app := NewApplication()
	app.Name = name

	// NEW: look into the package file (kind of fallback)
	for _, p := range m.Packages {
		for _, file := range p.Files {
			if file == name {
				pkg = p
				break
			}
		}
	}

	if pkg == nil {
		pkg = NewPackage()
		m.Packages = append(m.Packages, pkg)
	}

	pkg.Applications = append(pkg.Applications, app)
	return app, true
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
	app := NewApplication()
	app.Endpoints = append(app.Endpoints, &endpoint)
	// app := Application{Endpoints: []*ApplicationEndpoint{&endpoint}}

	pkg := NewPackage()
	pkg.Applications = append(pkg.Applications, app)
	// pkg := Package{Applications: []*Application{&app}}
	m.Packages = append(m.Packages, pkg)
	// m.Applications = append(m.Applications,
	// 	&Application{Endpoints: []*ApplicationEndpoint{&endpoint}})
	return app, true
}

// GetApplicationByPID returns a local app given its processus ID
func (m *Machine) GetApplicationByPID(pid uint) *Application {
	for _, p := range m.Applications() {
		if p.PID == pid {
			return p
		}
	}
	return nil
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
