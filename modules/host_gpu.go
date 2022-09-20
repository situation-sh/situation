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

	gpu, err := ghw.GPU()
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
		machine.GPUs = append(machine.GPUs, &g)
		l.Info("Found GPU on host")
	}
	return nil
}
