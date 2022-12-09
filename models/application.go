package models

import (
	"net"
)

// Package is a wrapper around application that stores distribution
// information of applications (executables)
type Package struct {
	Name            string         `json:"name,omitempty" jsonschema:"description=name of the package,example=openssh,example=musl-gcc,example=python,example=texlive"`
	Version         string         `json:"version,omitempty" jsonschema:"description=version of the package,example=8.8p1,example=1.2.3,example=3.10.8,example=2021"`
	Vendor          string         `json:"vendor,omitempty" jsonschema:"description=name of the organization that produce the package or its maintainer,example=Fedora Project,example=bob@debian.org"`
	Manager         string         `json:"manager,omitempty" jsonschema:"description=program that manages the package installation,example=rpm,example=dpkg,example=msi,example=builtin"`
	InstallTimeUnix int64          `json:"install_time,omitempty" jsonschema:"description=UNIX timestamp of the package installation,example=1670520587"`
	Applications    []*Application `json:"applications" jsonschema:"description=list of the applications provided by this package"`
	Files           []string       `json:"-"` // ignore that field for the moment
}

func NewPackage() *Package {
	return &Package{
		Applications: make([]*Application, 0),
		Files:        make([]string, 0),
	}
}

// Equal check if two packages are the same. Here we assume
// that Name and Version must be set
func (pkg *Package) Equal(other *Package) bool {
	if !(pkg.Name == other.Name && len(pkg.Name) > 0) {
		return false
	}
	if !(pkg.Version == other.Version && len(pkg.Version) > 0) {
		return false
	}
	if pkg.Vendor != other.Vendor {
		return false
	}
	if pkg.Manager != other.Manager {
		return false
	}
	return true
}

// ApplicationNames return the names of the apps that are
// attached to the package
func (pkg *Package) ApplicationNames() []string {
	if len(pkg.Applications) == 0 {
		return []string{}
	}
	out := make([]string, len(pkg.Applications))
	for i, a := range pkg.Applications {
		out[i] = a.Name
	}
	return out
}

// Application is a structure that represents all the
// types of apps we can have on a system
type Application struct {
	Name      string                 `json:"name,omitempty" jsonschema:"description=path (or name) of the application,example=/usr/sbin/sshd,example=/usr/bin/musl-gcc,example=C:\\Windows\\System32\\svchost.exe,example=wininit.exe,example=System"`
	Args      []string               `json:"args,omitempty" jsonschema:"description=list of arguments passed to app"` // we cannot put example right now (PR in progress: https://github.com/invopop/jsonschema/pull/31)
	Endpoints []*ApplicationEndpoint `json:"endpoints"  jsonschema:"description=list of network endpoints open by this app"`
}

// ApplicationEndpoint is a structure used by Application
// to tell that the app listens on given addr and port
type ApplicationEndpoint struct {
	Port     uint16 `json:"port" jsonschema:"description=port,example=22,example=80,example=443,example=49667"`
	Protocol string `json:"protocol" jsonschema:"description=transport layer protocol,example=tcp,example=udp"`
	Addr     net.IP `json:"addr" jsonschema:"description=binding IP address,example=0.0.0.0,example=::,example=127.0.0.1,example=192.168.122.23"`
}

func (s *Application) lastEndpoint() *ApplicationEndpoint {
	if len(s.Endpoints) == 0 {
		return nil
	}
	return s.Endpoints[len(s.Endpoints)-1]
}

// AddEndpoint appends a new endpoint if it does exist yet
// It returns true if a new endpoint has been added
func (s *Application) AddEndpoint(addr net.IP, port uint16, proto string) (*ApplicationEndpoint, bool) {
	// check if it exist
	for _, e := range s.Endpoints {
		if e.Addr.Equal(addr) && e.Port == port && e.Protocol == proto {
			// fmt.Println("Already got:", e)
			return e, false
		}
	}

	s.Endpoints = append(s.Endpoints,
		&ApplicationEndpoint{Addr: addr, Port: port, Protocol: proto})

	return s.lastEndpoint(), true
}
