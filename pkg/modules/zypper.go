//go:build linux

// LINUX(ZypperModule) ok
// WINDOWS(ZypperModule) no
// MACOS(ZypperModule) no
// ROOT(ZypperModule) no
package modules

import (
	"context"

	rpmdb "github.com/knqyf263/go-rpmdb/pkg"
	"github.com/situation-sh/situation/pkg/models"
)

const (
	zypperDefaultPath = "/var/lib/rpm/Packages.db"
	ZYPPER
)

var ZYPPER_BASED_FAMILIES = []string{"suse"}

// ZypperModule reads package information from the zypper package manager.
//
// This module is relevant for distros that use zypper, like suse and their
// derivatives. It uses [go-rpmdb].
//
// It reads `/var/lib/rpm/Packages.db`.
//
// [go-rpmdb]: https://github.com/knqyf263/go-rpmdb/
type ZypperModule struct {
	BaseModule
}

func init() {
	registerModule(&ZypperModule{})
}

func (m *ZypperModule) Name() string {
	return "zypper"
}

func (m *ZypperModule) Dependencies() []string {
	// depends on ping to ensure a rather fresh
	// arp table
	return []string{"host-basic", "netstat"}
}

func (m *ZypperModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	pm, err := NewAbstractPackageManager(ctx, ZYPPER_BASED_FAMILIES, logger, storage)
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

func parseZypperPackage(pkg *rpmdb.PackageInfo) *models.Package {
	p := models.Package{
		Name:            pkg.Name,
		Version:         pkg.Version,
		Vendor:          pkg.Vendor,
		Manager:         "zypper",
		InstallTimeUnix: int64(pkg.EpochNum()),
	}

	files, err := pkg.InstalledFiles()
	if err != nil {
		return &p
	}

	p.Files = make([]string, len(files))
	for i, file := range files {
		p.Files[i] = file.Path
	}

	return &p
}

func (m *ZypperModule) packageGenerator() (<-chan *models.Package, error) {
	db, err := rpmdb.Open(zypperDefaultPath)
	if err != nil {
		return nil, err
	}

	pkgs, err := db.ListPackages()
	if err != nil {
		return nil, err
	}

	c := make(chan *models.Package)

	go func() {
		for _, pkg := range pkgs {
			p := parseZypperPackage(pkg)
			c <- p
		}
		close(c)
	}()

	return c, nil
}

// machine := m.store.GetHost()
// if !utils.Includes([]string{"sles", "opensuse-leap", "opensuse-tumbleweed"}, machine.Distribution) {
// 	msg := fmt.Sprintf("The distribution %s is not supported", machine.Distribution)
// 	m.logger.Warn(msg)
// 	return &notApplicableError{msg: msg}
// }
// host := storage.GetOrCreateHost(ctx)
// if host == nil || host.ID == 0 {
// 	return fmt.Errorf("no host found in storage")
// }

// if host.DistributionFamily != "" && !utils.Includes(ZYPPER_BASED_FAMILIES, host.DistributionFamily) {
// 	logger.
// 		WithField("distribution_family", host.DistributionFamily).
// 		Warn("Module skipped for this distribution")
// 	return nil
// }

// appsToUpdate := make([]*models.Application, 0)
// fileAppMap, err := storage.BuildFileAppMap(ctx)
// if err != nil {
// 	// only warn here
// 	logger.Warn(err)
// }

// db, err := rpmdb.Open(zypperDefaultPath)
// if err != nil {
// 	return err
// }

// pkgs, err := db.ListPackages()
// if err != nil {
// 	return err
// }

// all := make([]*models.Package, 0)
// for _, pkg := range pkgs {
// 	p := parseZypperPackage(pkg)
// 	logger.WithField("name", p.Name).
// 		WithField("version", p.Version).
// 		WithField("install", time.Unix(p.InstallTimeUnix, 0)).
// 		WithField("files", len(p.Files)).
// 		Debug("Package found")

// 	// look at the files and update apps accordingly
// 	for _, f := range p.Files {
// 		if _, exists := fileAppMap[f]; exists {
// 			for _, app := range fileAppMap[f] {
// 				// link package to app (not the ID yet)
// 				app.Package = p
// 				appsToUpdate = append(appsToUpdate, app)
// 				logger.
// 					WithField("app", app.Name).
// 					WithField("file", f).
// 					Debug("Linking application to package")
// 			}
// 		}
// 	}

// 	all = append(all, p)
// }

// err = storage.InsertPackages(ctx, pkgs)
// if err != nil {
// 	return err
// }

// if len(appsToUpdate) > 0 {
// 	// update apps ID
// 	for _, app := range appsToUpdate {
// 		if app.Package != nil && app.Package.ID != 0 {
// 			app.PackageID = app.Package.ID
// 		}
// 	}

// 	_, err = storage.DB().
// 		NewUpdate().
// 		Model(&appsToUpdate).
// 		Column("package_id").
// 		Bulk().
// 		Exec(ctx)

// 	logger.
// 		WithField("apps", len(appsToUpdate)).
// 		Info("Applications linked to packages")

// 	return err
// }

// return nil
// }
