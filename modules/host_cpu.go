// LINUX(HostCPUModule) ok
// WINDOWS(HostCPUModule) ok
// MACOS(HostCPUModule) ?
// ROOT(HostCPUModule) no
package modules

import (
	"fmt"
	"strconv"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/store"
)

func init() {
	RegisterModule(&HostCPUModule{})
}

// Module definition ---------------------------------------------------------

// HostCPUModule retrieves host CPU info: model, vendor and
// the number of cores.
//
// It heavily relies on the [gopsutil] library.
//
// On Linux, it reads `/proc/cpuinfo`.
// On Windows it performs the `win32_Processor` WMI request
//
// On windows 11, the local user account must have administrator permissions (it does not mean it must be run as root).
//
// [gopsutil]: https://github.com/shirou/gopsutil/
type HostCPUModule struct{}

func (m *HostCPUModule) Name() string {
	return "host-cpu"
}

func (m *HostCPUModule) Dependencies() []string {
	return []string{"host-basic"}
}

func (m *HostCPUModule) Run() error {
	logger := GetLogger(m)
	machine := store.GetHost()
	if machine == nil {
		return fmt.Errorf("cannot retrieve host machine")
	}

	info, err := cpu.Info()
	if err != nil {
		return fmt.Errorf("error while retrieving CPU information: %v", err)
	}

	machine.CPU = &models.CPU{}
	machine.CPU.ModelName = info[0].ModelName
	machine.CPU.Vendor = info[0].VendorID
	logger.WithField(
		"model_name", machine.CPU.ModelName).WithField(
		"vendor", machine.CPU.Vendor).Info("Got CPU info on host")

	lastCoreID, err := strconv.Atoi(info[len(info)-1].CoreID)
	if err == nil {
		machine.CPU.Cores = lastCoreID + 1
		logger.WithField("cores", machine.CPU.Cores).Info("Get the number of cores")
		return nil
	}
	// fallback to the number of InfoStats
	machine.CPU.Cores = len(info)
	logger.WithField("cores", machine.CPU.Cores).Warn("Falling back to the number of records")
	logger.Errorf("cannot parse coreID: %s (%v)", info[len(info)-1].CoreID, err)
	return nil
}
