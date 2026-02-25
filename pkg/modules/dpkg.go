//go:build linux

// LINUX(DPKGModule) ok
// WINDOWS(DPKGModule) no
// MACOS(DPKGModule) no
// ROOT(DPKGModule) no
package modules

import (
	"context"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/modules/dpkg"
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
	return []string{"host-basic", "netstat"}
}

func (m *DPKGModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	pm, err := NewAbstractPackageManager(ctx, DPKG_BASED_FAMILIES, logger, storage)
	switch e := err.(type) {
	case nil:
		// continue
	case *notApplicableError:
		// module not applicable, skip
		return nil
	default:
		return e
	}

	generator, err := m.packageGenerator()
	if err != nil {
		return err
	}

	return pm.Run(generator)
}

// 	host := storage.GetOrCreateHost(ctx)
// 	if host == nil || host.ID == 0 {
// 		return fmt.Errorf("no host found in storage")
// 	}

// 	if host.DistributionFamily != "" && !utils.Includes(DPKG_BASED_FAMILIES, host.DistributionFamily) {
// 		logger.
// 			WithField("distribution_family", host.DistributionFamily).
// 			Warn("Module skipped for this distribution")
// 		return nil
// 	}

// 	appsToUpdate := make([]*models.Application, 0)
// 	fileAppMap, err := storage.BuildFileAppMap(ctx)
// 	if err != nil {
// 		// only warn here
// 		logger.Warn(err)
// 	}

// 	packages, err := dpkg.GetInstalledPackages()
// 	if err != nil {
// 		return err
// 	}
// 	for _, p := range packages {
// 		if len(p.Name) > 0 {
// 			p.MachineID = host.ID
// 			p.Files, err = dpkg.GetFiles(p.Name)
// 			logger.
// 				WithField("name", p.Name).
// 				WithField("version", p.Version).
// 				WithField("install", time.Unix(p.InstallTimeUnix, 0)).
// 				WithField("files", len(p.Files)).
// 				Debug("Package found")

// 			// look at the files and update apps accordingly
// 			for _, f := range p.Files {
// 				if _, exists := fileAppMap[f]; exists {
// 					for _, app := range fileAppMap[f] {
// 						// link package to app (not the ID yet)
// 						app.Package = p
// 						appsToUpdate = append(appsToUpdate, app)
// 						logger.
// 							WithField("app", app.Name).
// 							WithField("file", f).
// 							Debug("Linking application to package")
// 					}
// 				}
// 			}

// 		}
// 	}

// 	err = storage.InsertPackages(ctx, packages)
// 	if err != nil {
// 		return err
// 	}

// 	if len(appsToUpdate) > 0 {
// 		// update apps ID
// 		for _, app := range appsToUpdate {
// 			if app.Package != nil && app.Package.ID != 0 {
// 				app.PackageID = app.Package.ID
// 			}
// 		}

// 		_, err = storage.DB().
// 			NewUpdate().
// 			Model(&appsToUpdate).
// 			Column("package_id").
// 			Bulk().
// 			Exec(ctx)

// 		logger.
// 			WithField("apps", len(appsToUpdate)).
// 			Info("Applications linked to packages")

// 		return err
// 	}

// 	return nil
// }

func (m *DPKGModule) packageGenerator() (<-chan *models.Package, error) {
	packages, err := dpkg.GetInstalledPackages()
	if err != nil {
		return nil, err
	}

	c := make(chan *models.Package)
	go func() {
		for _, p := range packages {
			c <- p
		}
		close(c)
	}()

	return c, nil
}
