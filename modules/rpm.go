//go:build linux
// +build linux

// LINUX(RPMModule) ok
// WINDOWS(RPMModule) no
// MACOS(RPMModule) no
// ROOT(RPMModule) no
package modules

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/situation-sh/situation/modules/rpm"
	"github.com/situation-sh/situation/store"
	"github.com/situation-sh/situation/utils"
)

func init() {
	RegisterModule(&RPMModule{})
}

// RPMModule reads package information from the rpm package manager.
//
// This module is relevant for distros that use rpm, like fedora, redhat and their
// derivatives. It uses an sqlite client because of the way rpm works.
//
// It tries to read the rpm database: `/var/lib/rpm/rpmdb.sqlite`. Otherwise, it will
// try to find the `rpmdb.sqlite` file inside `/usr/lib`.
type RPMModule struct{}

func (m *RPMModule) Name() string {
	return "rpm"
}

func (m *RPMModule) Dependencies() []string {
	// depends on ping to ensure a rather fresh
	// arp table
	return []string{"host-basic", "netstat"}
}

func (m *RPMModule) Run() error {
	logger := GetLogger(m)
	machine := store.GetHost()
	if !utils.Includes([]string{"fedora", "rocky", "centos", "redhat", "almalinux", "opensuse-leap", "opensuse-tumbleweed"}, machine.Distribution) {
		msg := fmt.Sprintf("The distribution %s is not supported", machine.Distribution)
		logger.Warnf(msg)
		return &notApplicableError{msg: msg}
	}

	file, err := rpm.FindDBFile()
	if err != nil {
		if utils.Includes([]string{"opensuse-leap", "opensuse-tumbleweed"}, machine.Distribution) {
			msg := fmt.Sprintf("No RPM DB file found on this %s distro: %v (skipping)", machine.Distribution, err)
			logger.Warnf(msg)
			return nil
		}
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

	pkg := rpm.Pkg{}
	ins := rpm.Install{}

	for pkgRows.Next() {
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

		r := logger.WithField("name", p.Name).
			WithField("version", p.Version).
			WithField("install", time.Unix(p.InstallTimeUnix, 0).Format(time.RFC822))
		// here we can have issues if the packages already exist
		// ex: if a blank package has been created for an app
		// For the mapping, we ought to find if the application
		// name is within the files of the package
		// InsertPackage tries to do this
		x, merged := machine.InsertPackage(p)
		if merged {
			r.WithField("apps", x.ApplicationNames()).
				Info("Package merged with already found apps")
		} else {
			r.Debug("Package found")
		}
	}

	if err = pkgRows.Close(); err != nil {
		return fmt.Errorf("error while closing query: %v", err)
	}

	if err = db.Close(); err != nil {
		return fmt.Errorf("error while closing db: %v", err)
	}

	return nil
}
