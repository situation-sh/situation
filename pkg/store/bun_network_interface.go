package store

import (
	"context"

	"github.com/situation-sh/situation/pkg/models"
)

// GetMachineNICs returns all network interfaces for a given machine.
func (s *BunStorage) GetMachineNICs(ctx context.Context, machineID int64) []*models.NetworkInterface {
	nics := make([]*models.NetworkInterface, 0)
	err := s.db.
		NewSelect().
		Model(&nics).
		Where("machine_id = ?", machineID).
		Relation("Machine").
		Relation("Subnetworks").
		Scan(ctx)
	if err != nil {
		s.onError(err)
		return nil
	}
	return nics
}

// GetNICByMAC returns a network interface by its MAC address.
func (s *BunStorage) GetNICByMAC(ctx context.Context, mac string) *models.NetworkInterface {
	var nic models.NetworkInterface
	err := s.db.
		NewSelect().
		Model(&nic).
		Where("mac = ?", mac).
		Relation("Machine").
		Relation("Subnetwork").
		Limit(1).
		Scan(ctx, &nic)
	if err != nil {
		s.onError(err)
		return nil
	}
	return &nic
}

// GetNICByMACOnSubnet returns a network interface by its MAC address on a specific subnet.
func (s *BunStorage) GetNICByMACOnSubnet(ctx context.Context, mac string, subnetID int64) *models.NetworkInterface {
	var nic models.NetworkInterface
	err := s.db.
		NewSelect().
		Model(&nic).
		Where("mac = ? OR "+s.ANY("ip"), mac).
		Where("EXISTS (SELECT 1 FROM network_interface_subnets WHERE network_interface_id = network_interface.id AND subnetwork_id = ?)", subnetID).
		Relation("Machine").
		Relation("Subnetworks").
		Limit(1).
		Scan(ctx)
	if err != nil {
		s.onError(err)
		return nil
	}
	return &nic
}

// GetNICByIPOnSubnet returns a network interface by its IP address on a specific subnet.
func (s *BunStorage) GetNICByIPOnSubnet(ctx context.Context, ip string, subnetID int64) *models.NetworkInterface {
	var nic models.NetworkInterface
	err := s.db.
		NewSelect().
		Model(&nic).
		Where(s.ANY("ip"), ip).
		Where("EXISTS (SELECT 1 FROM network_interface_subnets WHERE network_interface_id = network_interface.id AND subnetwork_id = ?)", subnetID).
		Relation("Machine").
		Relation("Subnetworks").
		Limit(1).
		Scan(ctx)
	if err != nil {
		s.onError(err)
		return nil
	}
	return &nic
}

// GetNICByMACOrIPOnSubnet returns a network interface by its MAC or IP address on a specific subnet.
func (s *BunStorage) GetNICByMACOrIPOnSubnet(ctx context.Context, mac string, ip string, subnetID int64) (*models.NetworkInterface, error) {
	var nic models.NetworkInterface
	err := s.db.
		NewSelect().
		Model(&nic).
		Where("mac = ? OR "+s.ANY("ip"), mac, ip).
		Where("EXISTS (SELECT 1 FROM network_interface_subnets WHERE network_interface_id = network_interface.id AND subnetwork_id = ?)", subnetID).
		Relation("Machine").
		Relation("Subnetworks").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &nic, err
}

// GetNICsByIPs returns all network interfaces that have at least one IP
// in the provided list
func (s *BunStorage) GetNICsByIPs(ctx context.Context, ips []string) ([]models.NetworkInterface, error) {
	if len(ips) == 0 {
		return nil, nil
	}

	nics := make([]models.NetworkInterface, 0)
	err := s.db.NewSelect().
		Model(&nics).
		Where(s.OVERLAP("ip"), s.ARRAY(ips)).
		Relation("Machine").
		Relation("Subnetworks").
		Scan(ctx)
	if err != nil {
		s.onError(err)
		return nil, err
	}

	return nics, nil
}

// GetNICsByIPsOnSubnet returns all network interfaces that have at least one IP
// in the provided list and belong to the specified subnet.
// Works with both PostgreSQL (jsonb) and SQLite (json).
func (s *BunStorage) GetNICsByIPsOnSubnet(ctx context.Context, ips []string, subnetID int64) ([]models.NetworkInterface, error) {
	if len(ips) == 0 {
		return nil, nil
	}

	nics := make([]models.NetworkInterface, 0)
	err := s.db.NewSelect().
		Model(&nics).
		Where(s.OVERLAP("ip"), s.ARRAY(ips)).
		Where("EXISTS (SELECT 1 FROM network_interface_subnets WHERE network_interface_id = network_interface.id AND subnetwork_id = ?)", subnetID).
		Relation("Machine").
		Relation("Subnetworks").
		Scan(ctx)
	if err != nil {
		s.onError(err)
		return nil, err
	}

	return nics, nil
}

// EnsureNoOrphanNICs creates a dummy machine for all NICs that
// are not linked to any machine.
func (s *BunStorage) EnsureNoOrphanNICs(ctx context.Context) error {
	orphanNICs := make([]*models.NetworkInterface, 0)
	err := s.db.NewSelect().
		Model((*models.NetworkInterface)(nil)).
		Where("machine_id IS NULL OR machine_id = 0").
		Scan(ctx, &orphanNICs)
	if err != nil {
		s.onError(err)
		return err
	}

	if len(orphanNICs) > 0 {
		// create the dummy machines
		machines := make([]models.Machine, len(orphanNICs))
		for i := range orphanNICs {
			machines[i] = models.Machine{}
		}
		_, err := s.db.NewInsert().
			Model(&machines).
			Exec(ctx)
		if err != nil {
			s.onError(err)
			return err
		}

		// link the NICs to the newly created machines (bulk update)
		for i, nic := range orphanNICs {
			nic.MachineID = machines[i].ID
		}
		_, err = s.db.NewUpdate().
			Model(&orphanNICs).
			Column("machine_id").
			Bulk().
			Exec(ctx)
		if err != nil {
			s.onError(err)
			return err
		}
	}

	return nil
}

// GetNeighorNICS returns all NICs that share subnets with the host machine,
// excluding the host's own NICs.
func (s *BunStorage) GetNeighorNICS(ctx context.Context) ([]*models.NetworkInterface, error) {
	hostID := s.GetHostID(ctx)
	nics := make([]*models.NetworkInterface, 0)
	// Subquery to get subnet IDs connected to the host
	hostSubnets := s.db.NewSelect().
		TableExpr("network_interface_subnets AS nis").
		Column("nis.subnetwork_id").
		Join("JOIN network_interfaces AS ni ON ni.id = nis.network_interface_id").
		Where("ni.machine_id = ?", hostID)

	// Subquery to get NIC IDs that share those subnets
	nicIDs := s.db.NewSelect().
		TableExpr("network_interface_subnets AS nis").
		Column("nis.network_interface_id").
		Where("nis.subnetwork_id IN (?)", hostSubnets)

	// Get NICs that share subnets with host, excluding host's own NICs
	err := s.db.
		NewSelect().
		Model(&nics).
		Where("id IN (?)", nicIDs).
		Where("machine_id IS NULL OR machine_id <> ?", hostID).
		Scan(ctx)
	return nics, err
}

func (s *BunStorage) GetHostNICs(ctx context.Context) []*models.NetworkInterface {
	hostID := s.GetHostID(ctx)
	nics := make([]*models.NetworkInterface, 0)
	err := s.db.
		NewSelect().
		Model(&nics).
		Where("machine_id = ?", hostID).
		Relation("Machine").
		Relation("Subnetworks").
		Scan(ctx)
	if err != nil {
		s.onError(err)
		return nil
	}
	return nics
}

func (s *BunStorage) GetNICByMACAndIPs(ctx context.Context, mac string, ips []string) []*models.NetworkInterface {
	nics := make([]*models.NetworkInterface, 0)
	err := s.db.
		NewSelect().
		Model(&nics).
		Where("mac = ?", mac).
		Where(s.OVERLAP("ip"), s.ARRAY(ips)).
		Relation("Machine").
		Relation("Subnetworks").
		Scan(ctx)
	if err != nil {
		s.onError(err)
		return nil
	}
	return nics
}
