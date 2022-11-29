package models

import (
	"net"
)

// Package is a wrapper around application that stores distribution
// information of applications (executables)
type Package struct {
	Name            string         `json:"name,omitempty"`
	Version         string         `json:"version,omitempty"`
	Vendor          string         `json:"vendor,omitempty"`
	Manager         string         `json:"manager,omitempty"`
	InstallTimeUnix int64          `json:"install_time,omitempty"`
	Applications    []*Application `json:"applications"`
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

// Application is a structure that represents all the
// types of apps we can have on a system
type Application struct {
	Name      string                 `json:"name,omitempty"`
	Args      []string               `json:"args,omitempty"`
	Endpoints []*ApplicationEndpoint `json:"endpoints"`
}

// ApplicationEndpoint is a structure used by Application
// to tell that the app listens on given addr and port
type ApplicationEndpoint struct {
	Port     uint16 `json:"port"`
	Protocol string `json:"protocol"`
	Addr     net.IP `json:"addr"`
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
