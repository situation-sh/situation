//go:build linux

// LINUX(RPMModule) ok
// WINDOWS(RPMModule) no
// MACOS(RPMModule) no
// ROOT(RPMModule) no
package modules

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/modules/rpm"
	"github.com/situation-sh/situation/pkg/utils"
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
	return []string{"host-basic"}
}

func (m *RPMModule) Run(ctx context.Context) error {

	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	host := storage.GetOrCreateHost(ctx)
	if host == nil || host.ID == 0 {
		return fmt.Errorf("no host found in storage")
	}

	if host.DistributionFamily != "" && !utils.Includes(RPM_BASED_FAMILIES, host.DistributionFamily) {
		logger.
			WithField("distribution_family", host.DistributionFamily).
			Warn("Module skipped for this distribution")
		return nil
	}

	file, err := rpm.FindDBFile()
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite", file)
	if err != nil {
		return err
	}
	// defer db.Close()
	// db.SetConnMaxIdleTime(1 * time.Millisecond)
	// db.SetConnMaxLifetime(100 * time.Millisecond)

	// 1 connection for pkgRows
	// 1 connection for installRows
	// db.SetMaxOpenConns(2)

	pkgRows, err := db.Query("SELECT hnum, blob FROM Packages")
	// pkgRows, err := conn.QueryContext(ctx, "SELECT hnum, blob FROM Packages")
	if err != nil {
		return err
	}

	pkgs := make([]*models.Package, 0)

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
		p.MachineID = host.ID

		logger.WithField("name", p.Name).
			WithField("version", p.Version).
			WithField("install", time.Unix(p.InstallTimeUnix, 0)).
			WithField("files", len(p.Files)).
			Debug("Package found")

		pkgs = append(pkgs, p)
		// here we can have issues if the packages already exist
		// ex: if a blank package has been created for an app
		// For the mapping, we ought to find if the application
		// name is within the files of the package
		// InsertPackage tries to do this
		// x, merged := machine.InsertPackage(p)
		// if merged {
		// 	r.WithField("apps", x.ApplicationNames()).
		// 		Info("Package merged with already found apps")
		// } else {
		// 	r.Debug("Package found")
		// }
	}

	logger.
		WithField("packages", len(pkgs)).
		Info("RPM packages found")

	if err = pkgRows.Close(); err != nil {
		return fmt.Errorf("error while closing query: %v", err)
	}

	if err = db.Close(); err != nil {
		return fmt.Errorf("error while closing db: %v", err)
	}

	err = storage.InsertPackages(ctx, pkgs)
	return err
}
