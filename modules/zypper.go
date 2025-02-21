//go:build linux
// +build linux

// LINUX(ZypperModule) ok
// WINDOWS(ZypperModule) no
// MACOS(ZypperModule) no
// ROOT(ZypperModule) no
package modules

import (
	"fmt"
	"time"

	rpmdb "github.com/knqyf263/go-rpmdb/pkg"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/store"
	"github.com/situation-sh/situation/utils"
)

const (
	zypperDefaultPath = "/var/lib/rpm/Packages.db"
)

// ZypperModule reads package information from the zypper package manager.
//
// This module is relevant for distros that use zypper, like suse and their
// derivatives. It uses github.com/knqyf263/go-rpmdb/pkg.
//
// It reads `/var/lib/rpm/Packages.db`.
type ZypperModule struct{}

func init() {
	RegisterModule(&ZypperModule{})
}

func (m *ZypperModule) Name() string {
	return "zypper"
}

func (m *ZypperModule) Dependencies() []string {
	// depends on ping to ensure a rather fresh
	// arp table
	return []string{"host-basic", "netstat"}
}

func parseZypperPackage(pkg *rpmdb.PackageInfo) *models.Package {
	p := models.NewPackage()
	p.Name = pkg.Name
	p.Version = pkg.Version
	p.Vendor = pkg.Vendor
	p.Manager = "zypper"
	p.InstallTimeUnix = int64(pkg.EpochNum())

	files, err := pkg.InstalledFiles()
	if err != nil {
		return p
	}

	p.Files = make([]string, len(files))
	for i, file := range files {
		p.Files[i] = file.Path
	}

	return p
}

func (m *ZypperModule) Run() error {
	logger := GetLogger(m)
	machine := store.GetHost()
	if !utils.Includes([]string{"sles", "opensuse-leap", "opensuse-tumbleweed"}, machine.Distribution) {
		msg := fmt.Sprintf("The distribution %s is not supported", machine.Distribution)
		logger.Warn(msg)
		return &notApplicableError{msg: msg}
	}

	db, err := rpmdb.Open(zypperDefaultPath)
	if err != nil {
		return err
	}

	pkgs, err := db.ListPackages()
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		p := parseZypperPackage(pkg)
		r := logger.WithField("name", p.Name).
			WithField("version", p.Version).
			WithField("install", time.Unix(p.InstallTimeUnix, 0).Format(time.RFC822))
		x, merged := machine.InsertPackage(p)
		if merged {
			r.WithField("apps", x.ApplicationNames()).
				Info("Package merged with already found apps")
		} else {
			r.Debug("Package found")
		}
	}

	return nil
}
