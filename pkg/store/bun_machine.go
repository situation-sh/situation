package store

import (
	"context"
	"net"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/uptrace/bun"
)

// GetHost returns the host machine with all its related entities.
func (s *BunStorage) GetHost(ctx context.Context) *models.Machine {
	machine := new(models.Machine)
	err := s.db.NewSelect().
		Model(machine).
		Where("machine.agent = ?", s.agent).
		Relation("CPU").
		Relation("GPU").
		Relation("Disks").
		Relation("Packages").
		Relation("NICS").
		Scan(ctx)
	if err != nil {
		s.onError(err)
		return nil
	}
	return machine
}

// GetOrCreateHost returns the host machine or creates it if it doesn't exist.
func (s *BunStorage) GetOrCreateHost(ctx context.Context) *models.Machine {
	machine := models.Machine{Agent: s.agent}
	_, err := s.db.
		NewInsert().
		Model(&machine).
		On("CONFLICT (agent) DO UPDATE").
		Set("updated_at = CURRENT_TIMESTAMP").
		Returning("*").
		Exec(ctx)
	if err != nil {
		s.onError(err)
		return nil
	}
	s.cache.HostID = machine.ID
	return &machine
}

func (s *BunStorage) getHostIDFromCache(ctx context.Context) int64 {
	if s.cache.HostID != 0 {
		return s.cache.HostID
	}
	// Fallback to DB query
	s.GetOrCreateHost(ctx)
	return s.cache.HostID
}

// GetHostID returns the ID of the host machine.
func (s *BunStorage) GetHostID(ctx context.Context) int64 {
	return s.getHostIDFromCache(ctx)
}

// SetHostID manually sets the host ID in the cache.
// This is used by the fingerprint module when claiming an existing machine.
func (s *BunStorage) SetHostID(id int64) {
	s.cache.HostID = id
}

// GetorCreateHostCPU returns the CPU of the host machine or creates it if it doesn't exist.
func (s *BunStorage) GetorCreateHostCPU(ctx context.Context) *models.CPU {
	cpu := new(models.CPU)
	cpu.MachineID = s.getHostIDFromCache(ctx)

	err := s.db.NewInsert().
		Model(cpu).
		Ignore().
		Scan(ctx, cpu)

	if err != nil {
		s.onError(err)
		return nil
	}
	return cpu
}

// PreUpsertHost prepares an upsert query for a machine.
func (s *BunStorage) PreUpsertHost(m *models.Machine) *bun.InsertQuery {
	return s.db.
		NewInsert().
		Model(m).
		On("CONFLICT (agent) DO UPDATE").
		Set("updated_at = CURRENT_TIMESTAMP")
}

// PreUpdateMachine prepares an update query for a machine.
func (s *BunStorage) PreUpdateMachine(m *models.Machine) *bun.UpdateQuery {
	return s.db.
		NewUpdate().
		Model(m).
		Set("updated_at = CURRENT_TIMESTAMP")
}

// NewEmptyMachine creates a new empty machine in the database.
func (s *BunStorage) NewEmptyMachine(ctx context.Context) *models.Machine {
	m := new(models.Machine)
	if err := s.db.NewInsert().Model(m).Scan(ctx, m); err != nil {
		s.onError(err)
		return nil
	}
	return m
}

// IsHost returns true if the given machine is the host machine (the one running the agent).
func (s *BunStorage) IsHost(ctx context.Context, m *models.Machine) bool {
	return m.ID == s.GetHostID(ctx)
}

// GetMachineByIP returns the machine that has a NIC with the given IP address.
// It handles the multi-IP format (comma-separated IPs in NetworkInterface.IP).
// Returns nil if no machine is found.
func (s *BunStorage) GetMachineByIP(ctx context.Context, ip net.IP) *models.Machine {
	if ip == nil {
		return nil
	}
	ipStr := ip.String()
	machine := new(models.Machine)
	err := s.db.NewSelect().
		Model(machine).
		Join("JOIN network_interfaces ON network_interfaces.machine_id = machine.id").
		Where("network_interfaces.ip LIKE ?", "%"+ipStr+"%").
		Relation("NICS").
		Relation("Applications").
		Relation("Packages").
		Relation("Packages.Applications").
		Relation("Packages.Applications.Endpoints").
		Limit(1).
		Scan(ctx)
	if err != nil {
		s.onError(err)
		return nil
	}
	if machine.ID == 0 {
		return nil
	}
	return machine
}
