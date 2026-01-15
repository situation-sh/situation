// LINUX(HostCPUModule) ok
// WINDOWS(HostCPUModule) ok
// MACOS(HostCPUModule) ?
// ROOT(HostCPUModule) no
package modules

import (
	"context"
	"fmt"
	"strconv"

	"github.com/shirou/gopsutil/v4/cpu"
)

func init() {
	registerModule(&HostCPUModule{})
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
type HostCPUModule struct {
	BaseModule
}

func (m *HostCPUModule) Name() string {
	return "host-cpu"
}

func (m *HostCPUModule) Dependencies() []string {
	return []string{"host-basic"}
}

func (m *HostCPUModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	hostCPU := storage.GetorCreateHostCPU(ctx)
	if hostCPU == nil {
		return fmt.Errorf("unable to create or retrieve host CPU")
	}
	// machine := m.store.GetHost()
	// if machine == nil {
	// 	return fmt.Errorf("cannot retrieve host machine")
	// }

	info, err := cpu.Info()
	if err != nil {
		return fmt.Errorf("error while retrieving CPU information: %v", err)
	}

	query := storage.DB().NewUpdate().
		Model(hostCPU).
		Where("id = ?", hostCPU.ID).
		Set("model_name = ?", info[0].ModelName).
		Set("vendor = ?", info[0].VendorID)

	// hostCPU.ModelName = info[0].ModelName
	// hostCPU.Vendor = info[0].VendorID
	logger.
		WithField("model_name", hostCPU.ModelName).
		WithField("vendor", hostCPU.Vendor).
		Info("Got CPU info on host")

	lastCoreID, err := strconv.Atoi(info[len(info)-1].CoreID)
	if err == nil {
		// machine.CPU.Cores = lastCoreID + 1
		query = query.Set("cores = ?", lastCoreID+1)
		logger.WithField("cores", lastCoreID+1).Info("Get the number of cores")
	} else {
		// fallback to the number of InfoStats
		logger.WithField("cores", len(info)).Warn("Falling back to the number of records")
		query = query.Set("cores = ?", len(info))
	}

	_, err = query.Exec(ctx)

	return err
}
