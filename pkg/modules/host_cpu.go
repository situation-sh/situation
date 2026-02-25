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
	"github.com/situation-sh/situation/pkg/models"
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

	// hostCPU := storage.GetorCreateHostCPU(ctx)
	// if hostCPU == nil {
	// 	return fmt.Errorf("unable to create or retrieve host CPU")
	// }
	hostID := storage.GetHostID(ctx)
	if hostID == 0 {
		return fmt.Errorf("no host found in storage")
	}
	hcpu := models.CPU{
		MachineID: hostID,
	}
	// machine := m.store.GetHost()
	// if machine == nil {
	// 	return fmt.Errorf("cannot retrieve host machine")
	// }

	info, err := cpu.Info()
	if err != nil {
		return fmt.Errorf("error while retrieving CPU information: %v", err)
	}

	hcpu.ModelName = info[0].ModelName
	hcpu.Vendor = info[0].VendorID

	lastCoreID, err := strconv.Atoi(info[len(info)-1].CoreID)
	if err == nil {
		hcpu.Cores = lastCoreID + 1
	} else {
		// fallback to the number of InfoStats
		hcpu.Cores = len(info)
	}

	logger.
		WithField("model_name", hcpu.ModelName).
		WithField("vendor", hcpu.Vendor).
		WithField("cores", hcpu.Cores).
		WithField("host_id", hostID).
		Info("Got CPU info on host")

	_, err = storage.DB().
		NewInsert().
		Model(&hcpu).
		On("CONFLICT (machine_id) DO UPDATE").
		Set("model_name = EXCLUDED.model_name").
		Set("vendor = EXCLUDED.vendor").
		Set("cores = EXCLUDED.cores").
		Set("updated_at = CURRENT_TIMESTAMP").
		Exec(ctx)

	return err
}
