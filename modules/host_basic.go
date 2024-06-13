// LINUX(HostBasicModule) ok
// WINDOWS(HostBasicModule) ok
// MACOS(HostBasicModule) ?
// ROOT(HostBasicModule) no
package modules

import (
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/situation-sh/situation/config"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/store"
)

func init() {
	RegisterModule(&HostBasicModule{})
}

// Module definition ---------------------------------------------------------

// HostBasicModule retrieves basic information about the host:
// hostid, architecture, platform, distribution, version and uptime
//
// It heavily relies on the [gopsutil] library.
//
//	| Data                 | Linux                           | Windows                    |
//	|----------------------|---------------------------------|----------------------------|
//	| hostname             | `uname` syscall                 | `GetComputerNameExW` call  |
//	| arch                 | `uname` syscall                 | `GetNativeSystemInfo` call |
//	| platform             | `runtime.GOOS` variable         | `runtime.GOOS` variable    |
//	| distribution         | scanning `/etc/*-release` files | `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion\*` register keys |
//	| distribution version | scanning `/etc/*-release` files | `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion\*` register keys |
//	| hostid               | reading `/sys/class/dmi/id/product_uuid`, `/etc/machine-id` or `/proc/sys/kernel/random/boot_id` | `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography\MachineGuid` register key |
//	| uptime               | `sysinfo` syscall               | `GetTickCount64` call      |
//
// [gopsutil]: https://github.com/shirou/gopsutil/
type HostBasicModule struct{}

func (m *HostBasicModule) Name() string {
	return "host-basic"
}

func (m *HostBasicModule) Dependencies() []string {
	return nil
}

func (m *HostBasicModule) Run() error {
	logger := GetLogger(m)

	machine := models.NewMachine()
	if h, err := os.Hostname(); err == nil {
		machine.Hostname = h
		logger.WithField("hostname", machine.Hostname).Info("Get hostname")
	} else {
		logger.Errorf("Error while retrieving host hostname: %v", err)
	}

	if info, err := host.Info(); err == nil {
		machine.HostID = info.HostID
		machine.Arch = info.KernelArch
		machine.Platform = info.OS
		machine.Distribution = info.Platform
		machine.DistributionVersion = info.PlatformVersion
		// here the returned uptime is in seconds
		machine.Uptime = time.Duration(info.Uptime) * time.Second

		// logging
		entry := logger.WithField("arch", machine.Arch)
		entry = entry.WithField("platform", machine.Platform)
		entry = entry.WithField("distribution", machine.Distribution)
		entry = entry.WithField("distribution_version", machine.DistributionVersion)
		entry.Info("Get other Host infos")
	} else {
		logger.Errorf("Error while retrieving host infos: %v", err)
	}

	// config.SubConfig("").Print()
	// retrieve the agent from the config
	u, err := uuid.FromBytes(config.ID[:])
	if err != nil {
		logger.Error(err)
	}
	machine.Agent = &u
	logger.WithField("agent", machine.Agent).Info("Retrieve agent uuid")
	// insert the new machine!
	store.InsertMachine(machine)
	return nil
}
