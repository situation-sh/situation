//go:build linux
// +build linux

package modules

import (
	"fmt"

	"github.com/situation-sh/situation/modules/rpm"
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
	return []string{"host-basic"}
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

	pkgColl := session.Collection("Packages")
	installColl := session.Collection("Installtid")

	pkg := rpm.Pkg{}
	ins := rpm.Install{}
	res := pkgColl.Find()
	for res.Next(&pkg) {
		info := pkg.Parse()
		if err := installColl.Find(db.Cond{"hnum": pkg.Hnum}).One(&ins); err == nil {
			info["install"] = ins.Parse()
		}
		fmt.Println(info)
	}

	return nil
}
