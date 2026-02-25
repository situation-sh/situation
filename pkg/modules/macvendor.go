package modules

import (
	"context"
	"fmt"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/modules/macvendor"
)

func init() {
	registerModule(&MACVendorModule{})
}

// Module definition ---------------------------------------------------------

// MACVendorModule resolves manufacturer from MAC addresses.
//
// It uses a built-in lookup table of IEEE OUI assignments (generated
// from the official IEEE OUI registry) to match the first 3 octets
// of each MAC address to a vendor name.
//
// The module queries all network interfaces that have a MAC address
// but no vendor yet, and updates them in bulk.
type MACVendorModule struct {
	BaseModule
}

func (m *MACVendorModule) Name() string {
	return "macvendor"
}

func (m *MACVendorModule) Dependencies() []string {
	return []string{"arp"}
}

func (m *MACVendorModule) Run(ctx context.Context) error {
	storage := getStorage(ctx)
	logger := getLogger(ctx, m)

	interfaces := make([]*models.NetworkInterface, 0)
	err := storage.DB().
		NewSelect().
		Model(&interfaces).
		Where("mac_vendor IS NULL AND mac <> ''").
		Scan(ctx)
	if err != nil {
		return fmt.Errorf("macvendor: failed to retrieve network interfaces: %w", err)
	}

	toUpdate := make([]*models.NetworkInterface, 0)

	for _, iface := range interfaces {
		if vendor, exists := macvendor.Vendors[iface.MAC[:8]]; exists {
			iface.MACVendor = vendor
			toUpdate = append(toUpdate, iface)
		} else {
			logger.WithField("mac", iface.MAC).Debug("MAC vendor not found")
		}

	}

	if len(toUpdate) > 0 {
		_, err = storage.DB().
			NewUpdate().
			Model(&toUpdate).
			Column("mac_vendor").
			Bulk().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("macvendor: failed to update network interfaces: %w", err)
		}
		logger.WithField("network_interfaces", len(toUpdate)).Info("MAC vendors updated")
	} else {
		logger.Info("macvendor: no network interfaces to update")
	}

	return nil
}
