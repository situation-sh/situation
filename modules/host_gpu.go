// LINUX(HostGPUModule) ok
// WINDOWS(HostGPUModule) ok
// MACOS(HostGPUModule) ?
// ROOT(HostGPUModule) no
package modules

import (
	"fmt"

	"github.com/jaypipes/ghw"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/store"
)

func init() {
	RegisterModule(&HostGPUModule{})
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
type HostGPUModule struct{}

func (m *HostGPUModule) Name() string {
	return "host-gpu"
}

func (m *HostGPUModule) Dependencies() []string {
	return []string{"host-basic"}
}

func (m *HostGPUModule) Run() error {
	logger := GetLogger(m)
	machine := store.GetHost()
	if machine == nil {
		return fmt.Errorf("cannot retrieve host machine")
	}

	gpu, err := ghw.GPU(ghw.WithDisableWarnings())
	if err != nil {
		return fmt.Errorf("error while retrieving GPU information: %v", err)
	}

	for _, card := range gpu.GraphicsCards {
		// init structure
		g := models.GPU{
			Index: card.Index,
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
		machine.GPUS = append(machine.GPUS, &g)
		l.Info("Found GPU on host")
	}
	return nil
}
