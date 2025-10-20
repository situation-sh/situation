//go:build linux
// +build linux

// LINUX(DPKGModule) ok
// WINDOWS(DPKGModule) no
// MACOS(DPKGModule) no
// ROOT(DPKGModule) no
package modules

import (
	"fmt"
	"time"

	"github.com/situation-sh/situation/pkg/modules/dpkg"
	"github.com/situation-sh/situation/pkg/utils"
)

func init() {
	registerModule(&DPKGModule{})
}

// DPKGModule reads package information from the dpkg package manager.
//
// This module is relevant for distros that use dpkg, like debian, ubuntu and their
// derivatives. It only uses the standard library.
//
// It reads `/var/log/dpkg.log` and also files from `/var/lib/dpkg/info/`.
type DPKGModule struct {
	BaseModule
}

func (m *DPKGModule) Name() string {
	return "dpkg"
}

func (m *DPKGModule) Dependencies() []string {
	// host-basic is to check the distribution
	// netstat is to only fill the packages that have a running app
	// (see models.Machine.InsertPackages)
	return []string{"host-basic", "netstat"}
}

func (m *DPKGModule) Run() error {

	machine := m.store.GetHost()
	if !utils.Includes([]string{"debian", "ubuntu", "linuxmint", "raspbian"}, machine.Distribution) {
		msg := fmt.Sprintf("The distribution %s is not supported", machine.Distribution)
		m.logger.Warn(msg)
		return &notApplicableError{msg: msg}
	}
	packages, err := dpkg.GetInstalledPackages()
	if err != nil {
		return err
	}
	for _, p := range packages {
		if len(p.Name) > 0 {
			r := m.logger.WithField(
				"name", p.Name).WithField(
				"version", p.Version).WithField(
				"install", time.Unix(p.InstallTimeUnix, 0).Format(time.RFC822))

			p.Files, err = dpkg.GetFiles(p.Name)
			if err == nil {
				// add the package to the machine
				x, merged := machine.InsertPackage(p)
				if merged {
					r.WithField(
						"apps", x.ApplicationNames()).Info(
						"Package merged with already found apps")
				} else {
					r.Debug("Package found")
				}
			} else {
				r.Debug("Package found")
			}
		}
	}
	return nil
}
