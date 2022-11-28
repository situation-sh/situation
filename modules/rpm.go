//go:build linux
// +build linux

package modules

import (
	"github.com/situation-sh/situation/modules/rpm"
	"github.com/situation-sh/situation/store"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
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
	file, err := rpm.FindDBFile()
	if err != nil {
		return err
	}

	url, err := sqlite.ParseURL("file://" + file)
	if err != nil {
		return err
	}
	session, err := sqlite.Open(url)
	if err != nil {
		return err
	}
	defer session.Close()

	machine := store.GetHost()

	pkgColl := session.Collection("Packages")
	installColl := session.Collection("Installtid")

	pkg := rpm.Pkg{}
	ins := rpm.Install{}
	res := pkgColl.Find()
	for res.Next(&pkg) {
		p := pkg.Parse() // here we have a models.Package
		if err := installColl.Find(db.Cond{"hnum": pkg.Hnum}).One(&ins); err == nil {
			p.InstallTimeUnix = ins.Parse()
		}
		// here we can have issues if the packages already exist
		// ex: if a blank package has been created for an app
		// For the mapping, we ought to find if the application
		// name is within the files of the package
		// InsertPackage tries to do this
		machine.InsertPackage(p)
	}

	return nil
}
