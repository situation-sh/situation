// LINUX(HostBasicModule) ok
// WINDOWS(HostBasicModule) ok
// MACOS(HostBasicModule) ?
// ROOT(HostBasicModule) no
package modules

import (
	"os"
	"time"

	"github.com/shirou/gopsutil/v4/host"
	"github.com/situation-sh/situation/pkg/models"
)

func init() {
	registerModule(&HostBasicModule{})
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
type HostBasicModule struct {
	BaseModule
}

func (m *HostBasicModule) Name() string {
	return "host-basic"
}

func (m *HostBasicModule) Dependencies() []string {
	return nil
}

func (m *HostBasicModule) Run() error {

	machine := m.store.GetHost()
	if machine == nil {
		machine = models.NewMachine()
	}
	if h, err := os.Hostname(); err == nil {
		machine.Hostname = h
		m.logger.WithField("hostname", machine.Hostname).Info("Get hostname")
	} else {
		m.logger.Errorf("Error while retrieving host hostname: %v", err)
	}

	if info, err := host.Info(); err == nil {
		machine.HostID = info.HostID
		machine.Arch = info.KernelArch
		machine.Platform = info.OS
		machine.Distribution = info.Platform
		machine.DistributionVersion = info.PlatformVersion
		// here the returned uptime is in seconds
		if info.Uptime <= 0x7fffffffffffffff {
			machine.Uptime = time.Duration(info.Uptime) * time.Second
		}

		// logging
		entry := m.logger.WithField("arch", machine.Arch)
		entry = entry.WithField("platform", machine.Platform)
		entry = entry.WithField("distribution", machine.Distribution)
		entry = entry.WithField("distribution_version", machine.DistributionVersion)
		entry.Info("Get other Host infos")
	} else {
		m.logger.Errorf("Error while retrieving host infos: %v", err)
	}

	m.store.SetHost(machine)
	// // config.SubConfig("").Print()
	// // retrieve the agent from the config
	// u, err := uuid.FromBytes(config.ID[:])
	// if err != nil {
	// 	m.logger.Error(err)
	// }
	// machine.Agent = &u
	m.logger.WithField("agent", machine.Agent).Info("Retrieve agent uuid")
	// insert the new machine!
	m.store.InsertMachine(machine)
	return nil
}
