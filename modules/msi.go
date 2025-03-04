//go:build windows
// +build windows

// LINUX(MSIModule) no
// WINDOWS(MSIModule) ok
// MACOS(MSIModule) ?
// ROOT(MSIModule) yes
package modules

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/store"
	"golang.org/x/sys/windows/registry"
)

func init() {
	RegisterModule(&MSIModule{})
}

// MSIModule creates models.Packages instance from the windows registry
//
// For system-wide apps, it looks at `HKLM/SOFTWARE/Microsoft/Windows/CurrentVersion/Uninstall/*` and
// `HKLM/WOW6432Node/SOFTWARE/Microsoft/Windows/CurrentVersion/Uninstall/*` for 32bits apps.
// For user-specific apps: `HKCU/SOFTWARE/Microsoft/Windows/CurrentVersion/Uninstall/*`.
type MSIModule struct{}

func (m *MSIModule) Name() string {
	return "msi"
}

func (m *MSIModule) Dependencies() []string {
	// depends on ping to ensure a rather fresh
	// arp table
	return []string{"host-basic"}
}

func (m *MSIModule) Run() error {
	logger := GetLogger(m)

	host := store.GetHost()
	if host == nil {
		return fmt.Errorf("host not found")
	}
	systemApps, err := getInstalledApps(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`, logger)
	if err != nil {
		logger.Errorf("error fetching system-wide apps: %v", err)
	}
	systemApps32, err := getInstalledApps(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`, logger)
	if err != nil {
		logger.Errorf("error fetching 32-bit system-wide apps: %v", err)
	}
	userApps, err := getInstalledApps(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`, logger)
	if err != nil {
		logger.Errorf("error fetching user-specific apps: %v", err)
	}

	pkgs := append(systemApps, append(systemApps32, userApps...)...)
	for _, pkg := range pkgs {
		// InsertPackage does not append new package by default
		// It is more like a MergePackage
		out, merged := host.InsertPackage(pkg)
		if out == nil {
			// it means that the package has not been found
			// so has not been merged
			host.Packages = append(host.Packages, pkg)
			logger.WithField("name", pkg.Name).
				WithField("version", pkg.Version).
				WithField("vendor", pkg.Vendor).
				Info("Package created")
		} else if merged {
			logger.WithField("name", pkg.Name).
				WithField("version", pkg.Version).
				WithField("vendor", pkg.Vendor).
				Info("Package merged with a previous one")
		}
	}
	return nil
}

// findExecutables ignore root directory
func findExecutables(root string, maxDepth int) ([]string, error) {
	files := make([]string, 0)
	absPath, err := filepath.Abs(root)
	if err != nil {
		return files, err
	}
	if absPath == "/" || absPath == filepath.VolumeName(absPath)+`\` {
		return files, fmt.Errorf("root directory is not walked")
	}

	installLocation := os.DirFS(absPath)
	fs.WalkDir(installLocation, ".", func(path string, d fs.DirEntry, err error) error {
		// The err argument reports an error related to path,
		// signaling that WalkDir will not walk into that
		// directory.
		// The function can decide how to handle that error;
		// as described earlier, returning the error will cause
		// WalkDir to stop walking the entire tree.
		if err != nil {
			// continue
			return nil
		}
		// Calculate depth relative to root
		relPath, _ := filepath.Rel(root, path)
		depth := len(filepath.SplitList(relPath))

		if depth > maxDepth {
			// Skip deeper directories
			if d.IsDir() {
				return fs.SkipDir
			}
			// pass
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".exe") {
			files = append(files, filepath.Join(absPath, path))
		}
		return nil
	})

	return files, nil
}

// Get installed applications from Windows Registry
func getInstalledApps(root registry.Key, subKey string, logger *logrus.Entry) ([]*models.Package, error) {
	var err error
	pkgs := make([]*models.Package, 0)

	// Open registry key
	key, err := registry.OpenKey(root, subKey, registry.READ)
	if err != nil {
		return nil, err
	}
	defer key.Close()

	// Get subkeys (each app has its own entry)
	names, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	out := make(chan *models.Package, len(names))
	errs := make(chan error, len(names))

	// reader
	go func() {
		for pkg := range out {
			pkgs = append(pkgs, pkg)
		}
	}()

	// error reader
	go func() {
		for e := range errs {
			if err == nil {
				err = e
			}
		}
	}()

	for _, name := range names {
		subKeyPath := subKey + `\` + name
		logger.Debugf("Looking for registry key: %v", subKeyPath)

		wg.Add(1)
		// writer
		go func(subKeyPath string) {
			defer wg.Done()

			subKey, err := registry.OpenKey(root, subKeyPath, registry.READ)
			if err != nil {
				errs <- err
				return
			}
			defer subKey.Close()

			pkg := models.NewPackage()
			pkg.Manager = "msi"

			// Read string values
			if value, _, err := subKey.GetStringValue("DisplayName"); err == nil {
				pkg.Name = value
			}
			if value, _, err := subKey.GetStringValue("DisplayVersion"); err == nil {
				pkg.Version = value
			}
			if value, _, err := subKey.GetStringValue("Publisher"); err == nil {
				pkg.Vendor = value
			}
			if value, _, err := subKey.GetStringValue("InstallDate"); err == nil {
				if t, err := time.Parse("20060201", value); err == nil {
					pkg.InstallTimeUnix = t.Unix()
				}
			}
			if value, _, err := subKey.GetStringValue("InstallLocation"); err == nil {
				logger.Debugf("InstallLocation: %v", value)
				if files, err := findExecutables(value, 3); err == nil {
					pkg.Files = append(pkg.Files, files...)
				}
			}

			// check if system component
			isSystem := false
			if value, _, err := subKey.GetIntegerValue("SystemComponent"); err == nil {
				isSystem = value == 1
			}

			// Ignore system components if found
			if !isSystem && pkg.Name != "" {
				out <- pkg
			}
		}(subKeyPath)
	}

	wg.Wait()
	close(out)
	close(errs)

	return pkgs, err
}
