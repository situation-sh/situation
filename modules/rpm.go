//go:build linux
// +build linux

package modules

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"

	"github.com/situation-sh/situation/modules/rpm"
	"github.com/situation-sh/situation/store"
)

func init() {
	RegisterModule(&RPMModule{})
}

type RPMModule struct{}

func (m *RPMModule) Name() string {
	return "rpm"
}

func (m *RPMModule) Dependencies() []string {
	// depends on ping to ensure a rather fresh
	// arp table
	return []string{"netstat"}
}

func (m *RPMModule) Run() error {
	logger := GetLogger(m)
	file, err := rpm.FindDBFile()
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite", file)
	if err != nil {
		return err
	}

	pkgRows, err := db.Query("SELECT hnum, blob FROM Packages")
	if err != nil {
		return err
	}

	machine := store.GetHost()
	pkg := rpm.Pkg{}
	ins := rpm.Install{}
	for pkgRows.Next() {
		if err := pkgRows.Scan(&pkg.Hnum, &pkg.Blob); err != nil {
			continue
		}
		p := pkg.Parse() // here we have a models.Package
		installRows, err := db.Query("SELECT key, hnum, idx FROM Installtid WHERE hnum=? LIMIT 1", pkg.Hnum)
		if err != nil || installRows == nil {
			continue
		}
		if installRows.Next() {
			if err := installRows.Scan(&ins.Key, &ins.Hnum, &ins.Idx); err != nil {
				continue
			}
		}
		p.InstallTimeUnix = ins.Parse()

		r := logger.WithField(
			"name", p.Name).WithField(
			"version", p.Version).WithField(
			"install", time.Unix(p.InstallTimeUnix, 0).Format(time.RFC822))
		// here we can have issues if the packages already exist
		// ex: if a blank package has been created for an app
		// For the mapping, we ought to find if the application
		// name is within the files of the package
		// InsertPackage tries to do this
		x, merged := machine.InsertPackage(p)
		if merged {
			r.WithField(
				"apps", x.ApplicationNames()).Info(
				"Package merged with already found apps")
		} else {
			r.Debug("Package found")
		}
	}

	return nil
}
