// LINUX(AppUserModule) ok
// WINDOWS(AppUserModule) ok
// MACOS(AppUserModule) ?
// ROOT(AppUserModule) ?
package modules

import (
	"fmt"

	"github.com/situation-sh/situation/modules/appuser"
	"github.com/situation-sh/situation/store"
)

func init() {
	RegisterModule(&AppUserModule{})
}

// AppUserModule fills user information from the PID of an application
//
// On Linux, it uses the /proc/<PID>/status entrypoint.
// On Windows, it calls `OpenProcessToken`, `GetTokenInformation` and `LookupAccountSidW`.
//
// On windows, even if the agent is run as administrator, it may not have
// the required privileges to scan some processes like wininit.exe, services.exe.
type AppUserModule struct{}

func (m *AppUserModule) Name() string {
	return "appuser"
}

func (m *AppUserModule) Dependencies() []string {
	// depends on ping to ensure a rather fresh
	// arp table
	return []string{"netstat"}
}

func (m *AppUserModule) Run() error {
	logger := GetLogger(m)

	logger.Info("Filling user information on applications")

	host := store.GetHost()
	if host == nil {
		return fmt.Errorf("no host found")
	}

	for _, pkg := range host.Packages {
		for _, app := range pkg.Applications {
			if app.PID > 0 {
				if err := appuser.PopulateApplication(app); err != nil {
					logger.Error(err)
					continue
				}
			}
		}
	}

	return nil
}
