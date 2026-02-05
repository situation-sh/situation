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

	_ "modernc.org/sqlite"

	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/modules/rpm"
)

// see https://github.com/shirou/gopsutil/blob/master/host/host_linux.go#L215
var RPM_BASED_FAMILIES = []string{
	"fedora", "rhel",
	// "suse", ignore because of zypper
	"neokylin", "anolis",
}

var MIN_VERSION_SUPPORTED = map[string]string{
	"rockylinux":    "9",
	"opensuse-leap": "16",
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

	// extra checks
	if ignore, err := rpmMustBeIgnored(pm.host, logger); err != nil {
		return err
	} else if ignore {
		return nil
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

	db, err := sql.Open("sqlite", "file:"+file+"?mode=ro&immutable=1")
	if err != nil {
		return nil, fmt.Errorf("fail to open %v: %w", file, err)
	}

	pkgRows, err := db.Query("SELECT hnum, blob FROM Packages")
	// pkgRows, err := conn.QueryContext(ctx, "SELECT hnum, blob FROM Packages")
	if err != nil {
		return nil, fmt.Errorf("fail to query database: %w", err)
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

func rpmMustBeIgnored(m *models.Machine, logger logrus.FieldLogger) (bool, error) {
	minVersion, tocheck := MIN_VERSION_SUPPORTED[m.Distribution]
	if tocheck {
		vMin := version.Must(version.NewVersion(minVersion))
		hostVersion, err := version.NewVersion(m.DistributionVersion)
		if err != nil {
			return false, fmt.Errorf("cannot parse host distribution version: %w", err)
		}
		if hostVersion.LessThan(vMin) {
			// ignore the distribution if version is less than the minimum supported
			logger.
				WithField("distribution_version", m.DistributionVersion).
				WithField("distribution", m.Distribution).
				WithField("min_supported_version", minVersion).
				Info("Module skipped ")
			return true, nil
		}
	}
	return false, nil
}
