// LINUX(HostGPUModule) ok
// WINDOWS(HostGPUModule) ok
// MACOS(HostGPUModule) ?
// ROOT(HostGPUModule) no
package modules

import (
	"context"
	"fmt"

	"github.com/jaypipes/ghw"
	"github.com/situation-sh/situation/pkg/models"
)

func init() {
	registerModule(&HostGPUModule{})
}

// Module definition ---------------------------------------------------------

// HostGPUModule retrieves basic information about GPU:
// index, vendor and product name.
//
// It heavily relies on [ghw].
// On Linux it reads `/sys/class/drm/` folder. On Windows, it performs
// the following WMI query:
//
//	```ps1
//	SELECT Caption, CreationClassName, Description, DeviceID, Manufacturer, Name, PNPClass, PNPDeviceID FROM Win32_PnPEntity
//	```
//
// On windows 11, the local user account must have administrator permissions (it does not mean it must be run as root).
//
// [ghw]: https://github.com/jaypipes/ghw
type HostGPUModule struct {
	BaseModule
}

func (m *HostGPUModule) Name() string {
	return "host-gpu"
}

func (m *HostGPUModule) Dependencies() []string {
	return []string{"host-basic"}
}

func (m *HostGPUModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	hostID := storage.GetHostID(ctx)
	// machine := m.store.GetHost()
	// if machine == nil {
	// 	return fmt.Errorf("cannot retrieve host machine")
	// }

	gpus := make([]models.GPU, 0)
	gpu, err := ghw.GPU(ghw.WithDisableWarnings())
	if err != nil {
		return fmt.Errorf("error while retrieving GPU information: %v", err)
	}

	for _, card := range gpu.GraphicsCards {
		// init structure
		g := models.GPU{
			Index:     card.Index,
			MachineID: hostID,
		}
		l := logger.WithField("index", g.Index)

		dinfo := card.DeviceInfo
		// check sub structures to fill other fields
		if dinfo != nil {
			g.Driver = dinfo.Driver
			l = l.WithField("driver", g.Driver)
			if dinfo.Vendor != nil {
				g.Vendor = dinfo.Vendor.Name
				l = l.WithField("vendor", g.Vendor)
			}
			if dinfo.Product != nil {
				g.Product = dinfo.Product.Name
				l = l.WithField("product", g.Product)
			}
		}
		gpus = append(gpus, g)
		l.Info("Found GPU on host")
	}

	// update DB
	// we update everything even if input data may depend on the
	// content of dinfo on every card. This is not ideal but
	// it seems reliable enough for now.
	_, err = storage.DB().NewInsert().
		Model(&gpus).
		On("CONFLICT (machine_id, index) DO UPDATE").
		Set("product = EXCLUDED.product").
		Set("vendor = EXCLUDED.vendor").
		Set("driver = EXCLUDED.driver").
		Set("updated_at = CURRENT_TIMESTAMP").
		Exec(ctx)
	return nil
}
