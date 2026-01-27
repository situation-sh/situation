//go:build linux

// LINUX(DPKGModule) ok
// WINDOWS(DPKGModule) no
// MACOS(DPKGModule) no
// ROOT(DPKGModule) no
package modules

import (
	"context"
	"fmt"
	"time"

	"github.com/situation-sh/situation/pkg/modules/dpkg"
	"github.com/situation-sh/situation/pkg/utils"
)

// see https://github.com/shirou/gopsutil/blob/master/host/host_linux.go#L215
var DPKG_BASED_FAMILIES = []string{
	"debian",
}

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
	return []string{"host-basic"}
}

func (m *DPKGModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	host := storage.GetOrCreateHost(ctx)
	if host == nil || host.ID == 0 {
		return fmt.Errorf("no host found in storage")
	}

	if host.DistributionFamily != "" && !utils.Includes(DPKG_BASED_FAMILIES, host.DistributionFamily) {
		logger.
			WithField("distribution_family", host.DistributionFamily).
			Warn("Module skipped for this distribution")
		return nil
	}

	packages, err := dpkg.GetInstalledPackages()
	if err != nil {
		return err
	}
	for _, p := range packages {
		if len(p.Name) > 0 {
			p.MachineID = host.ID
			p.Files, err = dpkg.GetFiles(p.Name)
			logger.
				WithField("name", p.Name).
				WithField("version", p.Version).
				WithField("install", time.Unix(p.InstallTimeUnix, 0)).
				WithField("files", len(p.Files)).
				Info("Package found")

			// populate files

		}
	}

	err = storage.InsertPackages(ctx, packages)
	return err
}
