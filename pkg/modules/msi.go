//go:build windows

// LINUX(MSIModule) no
// WINDOWS(MSIModule) ok
// MACOS(MSIModule) ?
// ROOT(MSIModule) yes
package modules

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/pkg/models"
	"golang.org/x/sys/windows/registry"
)

func init() {
	registerModule(&MSIModule{})
}

// MSIModule creates models.Packages instance from the windows registry
//
// For system-wide apps, it looks at `HKLM/SOFTWARE/Microsoft/Windows/CurrentVersion/Uninstall/*` and
// `HKLM/WOW6432Node/SOFTWARE/Microsoft/Windows/CurrentVersion/Uninstall/*` for 32bits apps.
// For user-specific apps: `HKCU/SOFTWARE/Microsoft/Windows/CurrentVersion/Uninstall/*`.
type MSIModule struct {
	BaseModule
}

func (m *MSIModule) Name() string {
	return "msi"
}

func (m *MSIModule) Dependencies() []string {
	// depends on ping to ensure a rather fresh
	// arp table
	return []string{"host-basic"}
}

func (m *MSIModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)
	host := storage.GetOrCreateHost(ctx)
	if host == nil || host.ID == 0 {
		return fmt.Errorf("no host found in storage")
	}
	systemApps, err := getInstalledApps(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`, logger)
	if err != nil {
		logger.
			WithField("key", `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`).
			WithError(err).
			Error("error fetching system-wide apps")
	}
	systemApps32, err := getInstalledApps(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`, logger)
	if err != nil {
		logger.
			WithField("key", `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`).
			WithError(err).
			Error("error fetching 32-bit system-wide apps")
	}
	userApps, err := getInstalledApps(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`, logger)
	if err != nil {
		logger.
			WithField("key", `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`).
			WithError(err).
			Error("error fetching user-specific apps")
	}

	pkgs := append(systemApps, append(systemApps32, userApps...)...)

	err = storage.InsertPackages(ctx, pkgs)
	return err
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
func getInstalledApps(root registry.Key, subKey string, logger logrus.FieldLogger) ([]*models.Package, error) {
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
	// errs := make(chan error, len(names))

	for _, name := range names {
		subKeyPath := subKey + `\` + name
		logger.Debugf("Looking for registry key: %v", subKeyPath)

		wg.Add(1)
		go func(subKeyPath string) {
			defer wg.Done()

			subKey, err := registry.OpenKey(root, subKeyPath, registry.READ)
			if err != nil {
				logger.
					WithError(err).
					WithField("subkey", subKeyPath).
					Warn("Cannot open registry subkey")
				// errs <- err
				return
			}
			defer subKey.Close()

			pkg := &models.Package{
				Manager: "msi",
			}

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

	// Wait for all workers to finish, then close channels
	wg.Wait()
	close(out)

	// Collect results synchronously (channels are closed, safe to drain)
	pkgs := make([]*models.Package, 0, len(out))
	for pkg := range out {
		logger.
			WithField("name", pkg.Name).
			WithField("version", pkg.Version).
			WithField("install", time.Unix(pkg.InstallTimeUnix, 0)).
			WithField("files", len(pkg.Files)).
			Info("Package found")
		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}
