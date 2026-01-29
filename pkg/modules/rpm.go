//go:build linux

// LINUX(RPMModule) ok
// WINDOWS(RPMModule) no
// MACOS(RPMModule) no
// ROOT(RPMModule) no
package modules

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/modules/rpm"
)

// see https://github.com/shirou/gopsutil/blob/master/host/host_linux.go#L215
var RPM_BASED_FAMILIES = []string{
	"fedora", "rhel", "suse",
	"neokylin", "anolis",
}

func init() {
	registerModule(&RPMModule{})
}

// RPMModule reads package information from the rpm package manager.
//
// This module is relevant for distros that use rpm, like fedora, redhat and their
// derivatives. It uses an sqlite client because of the way rpm works.
//
// It tries to read the rpm database: `/var/lib/rpm/rpmdb.sqlite`. Otherwise, it will
// try to find the `rpmdb.sqlite` file inside `/usr/lib`.
type RPMModule struct {
	BaseModule
}

func (m *RPMModule) Name() string {
	return "rpm"
}

func (m *RPMModule) Dependencies() []string {
	// depends on ping to ensure a rather fresh
	// arp table
	return []string{"host-basic", "netstat"}
}

func (m *RPMModule) Run(ctx context.Context) error {

	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	pm, err := NewAbstractPackageManager(ctx, RPM_BASED_FAMILIES, logger, storage)
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

func (m *RPMModule) packageGenerator() (<-chan *models.Package, error) {
	// Implementation of the package generator goes here
	file, err := rpm.FindDBFile()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", file)
	if err != nil {
		return nil, err
	}

	pkgRows, err := db.Query("SELECT hnum, blob FROM Packages")
	// pkgRows, err := conn.QueryContext(ctx, "SELECT hnum, blob FROM Packages")
	if err != nil {
		return nil, err
	}

	c := make(chan *models.Package)

	go func() {
		defer close(c)
		defer pkgRows.Close()
		defer db.Close()
		for pkgRows.Next() {
			pkg := rpm.Pkg{}
			ins := rpm.Install{}

			// fmt.Printf("%+v\n", db.Stats())
			if err := pkgRows.Scan(&pkg.Hnum, &pkg.Blob); err != nil {
				continue
			}
			p := pkg.Parse() // here we have a models.Package
			installRows, err := db.Query("SELECT key, hnum, idx FROM Installtid WHERE hnum=? LIMIT 1", pkg.Hnum)
			// installRows, err := conn.QueryContext(ctx, "SELECT key, hnum, idx FROM Installtid WHERE hnum=? LIMIT 1", pkg.Hnum)
			if err != nil || installRows == nil {
				continue
			}
			if installRows.Next() {
				if err := installRows.Scan(&ins.Key, &ins.Hnum, &ins.Idx); err != nil {
					continue
				}
			}
			if err := installRows.Close(); err != nil {
				// once again ignore on error
				continue
			}
			p.InstallTimeUnix = ins.Parse()
			c <- p
		}
	}()

	return c, nil
}

// host := storage.GetOrCreateHost(ctx)
// if host == nil || host.ID == 0 {
// 	return fmt.Errorf("no host found in storage")
// }

// if host.DistributionFamily != "" && !utils.Includes(RPM_BASED_FAMILIES, host.DistributionFamily) {
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

// file, err := rpm.FindDBFile()
// if err != nil {
// 	return err
// }

// db, err := sql.Open("sqlite", file)
// if err != nil {
// 	return err
// }
// // defer db.Close()
// // db.SetConnMaxIdleTime(1 * time.Millisecond)
// // db.SetConnMaxLifetime(100 * time.Millisecond)

// // 1 connection for pkgRows
// // 1 connection for installRows
// // db.SetMaxOpenConns(2)

// pkgRows, err := db.Query("SELECT hnum, blob FROM Packages")
// // pkgRows, err := conn.QueryContext(ctx, "SELECT hnum, blob FROM Packages")
// if err != nil {
// 	return err
// }

// pkgs := make([]*models.Package, 0)

// for pkgRows.Next() {
// 	pkg := rpm.Pkg{}
// 	ins := rpm.Install{}

// 	// fmt.Printf("%+v\n", db.Stats())
// 	if err := pkgRows.Scan(&pkg.Hnum, &pkg.Blob); err != nil {
// 		continue
// 	}
// 	p := pkg.Parse() // here we have a models.Package
// 	installRows, err := db.Query("SELECT key, hnum, idx FROM Installtid WHERE hnum=? LIMIT 1", pkg.Hnum)
// 	// installRows, err := conn.QueryContext(ctx, "SELECT key, hnum, idx FROM Installtid WHERE hnum=? LIMIT 1", pkg.Hnum)
// 	if err != nil || installRows == nil {
// 		continue
// 	}
// 	if installRows.Next() {
// 		if err := installRows.Scan(&ins.Key, &ins.Hnum, &ins.Idx); err != nil {
// 			continue
// 		}
// 	}
// 	if err := installRows.Close(); err != nil {
// 		// once again ignore on error
// 		continue
// 	}
// 	p.InstallTimeUnix = ins.Parse()
// 	p.MachineID = host.ID

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

// 	// append
// 	pkgs = append(pkgs, p)
// }

// logger.
// 	WithField("packages", len(pkgs)).
// 	Info("RPM packages found")

// if err = pkgRows.Close(); err != nil {
// 	return fmt.Errorf("error while closing query: %v", err)
// }

// if err = db.Close(); err != nil {
// 	return fmt.Errorf("error while closing db: %v", err)
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
