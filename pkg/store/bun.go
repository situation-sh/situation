package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/utils"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
)

type Cache struct {
	HostID int64 // ID of the host machine in the db
}

type BunStorage struct {
	db      *bun.DB
	agent   string
	onError func(error)
	cache   Cache
}

func newStorage(db *bun.DB, agent string, onError func(error)) *BunStorage {
	// register m2m models
	db.RegisterModel((*models.ApplicationEndpoint)(nil))
	db.RegisterModel((*models.UserApplication)(nil))
	db.RegisterModel((*models.NetworkInterfaceSubnet)(nil))

	// fallback onError handler
	if onError == nil {
		onError = func(err error) {}
	}

	return &BunStorage{
		db:      db,
		agent:   agent,
		onError: onError,
		cache:   Cache{HostID: -1},
	}
}

func NewStorage(dataSourceName string, agent string, onError func(error)) (*BunStorage, error) {
	// Simple heuristic to choose between SQLite and Postgres
	switch detectDBType(dataSourceName) {
	case "sqlite":
		return NewSQLiteBunStorage(dataSourceName, agent, onError)
	case "postgres":
		return NewPostgresBunStorage(dataSourceName, agent, onError)
	default:
		return nil, fmt.Errorf("cannot detect database type with dsn=%s", dataSourceName)
	}
}

func NewSQLiteBunStorage(dataSourceName string, agent string, onError func(error)) (*BunStorage, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, dataSourceName)
	if err != nil {
		return nil, err
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())
	return newStorage(db, agent, onError), nil
}

func NewPostgresBunStorage(dataSourceName string, agent string, onError func(error)) (*BunStorage, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dataSourceName)))
	db := bun.NewDB(sqldb, pgdialect.New())
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return newStorage(db, agent, onError), nil
}

func detectDBType(dsn string) string {
	dsn = strings.TrimSpace(dsn)

	// PostgreSQL URL format
	if strings.HasPrefix(dsn, "postgres://") ||
		strings.HasPrefix(dsn, "postgresql://") {
		return "postgres"
	}

	// Check if it looks like a file path or SQLite DSN
	if strings.HasPrefix(dsn, "file:") ||
		strings.HasSuffix(dsn, ".db") ||
		strings.HasSuffix(dsn, ".sqlite") ||
		strings.HasSuffix(dsn, ".sqlite3") ||
		dsn == ":memory:" {
		return "sqlite"
	}

	// If it contains these patterns, likely PostgreSQL key-value format
	if strings.Contains(dsn, "host=") ||
		strings.Contains(dsn, "dbname=") ||
		strings.Contains(dsn, "user=") {
		return "postgres"
	}

	// Default to SQLite for simple paths
	if !strings.Contains(dsn, "://") {
		return "sqlite"
	}

	return ""
}

func (s *BunStorage) DB() *bun.DB {
	return s.db
}

func (s *BunStorage) Migrate(ctx context.Context) error {
	allModels := []interface{}{
		(*models.Subnetwork)(nil),
		(*models.Machine)(nil),
		(*models.CPU)(nil),
		(*models.GPU)(nil),
		(*models.Disk)(nil),
		(*models.NetworkInterface)(nil),
		(*models.NetworkInterfaceSubnet)(nil),
		(*models.Package)(nil),
		(*models.Application)(nil),
		(*models.ApplicationEndpoint)(nil),
		(*models.User)(nil),
		(*models.UserApplication)(nil),
	}

	for _, model := range allModels {
		_, err := s.db.NewCreateTable().
			Model(model).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}
	}

	return nil

}

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

func (s *BunStorage) GetorCreateHost(ctx context.Context) *models.Machine {
	machine := new(models.Machine)
	machine.Agent = s.agent
	err := s.db.
		NewInsert().
		Model(machine).
		Ignore().
		Scan(ctx, machine) // maybe machine is not important here since Scan seems to fill it automatically
	if err != nil {
		s.onError(err)
		return nil
	}
	s.cache.HostID = machine.ID
	return machine
}

func (s *BunStorage) getHostIDFromCache(ctx context.Context) int64 {
	if s.cache.HostID != 0 {
		return s.cache.HostID
	}
	// Fallback to DB query
	s.GetorCreateHost(ctx)
	return s.cache.HostID
}

func (s *BunStorage) GetHostID(ctx context.Context) int64 {
	return s.getHostIDFromCache(ctx)
}

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

// PreUpsert prepares an upsert query for a machine with
func (s *BunStorage) PreUpsertHost(m *models.Machine) *bun.InsertQuery {
	return s.db.
		NewInsert().
		Model(m).
		On("CONFLICT (agent) DO UPDATE").
		Set("updated_at = CURRENT_TIMESTAMP")
}

func (s *BunStorage) PreUpdateMachine(m *models.Machine) *bun.UpdateQuery {
	return s.db.
		NewUpdate().
		Model(m).
		Set("updated_at = CURRENT_TIMESTAMP")
}

func (s *BunStorage) GetMachineNICs(ctx context.Context, machineID int64) []models.NetworkInterface {
	// var model *models.NetworkInterface = nil
	nics := make([]models.NetworkInterface, 0)
	// err := s.db.
	// 	NewSelect().
	// 	Model((*models.NetworkInterface)(nil)).
	// 	Relation("Subnetworks").
	// 	Scan(ctx, &nics)

	_, err := s.db.
		NewSelect().
		Model((*models.NetworkInterface)(nil)).
		Where("machine_id = ?", machineID).
		Relation("Machine").
		Relation("Subnetworks").
		Exec(ctx)
	if err != nil {
		s.onError(err)
		return nil
	}
	return nics
}

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

func (s *BunStorage) GetNICByMACOnSubnet(ctx context.Context, mac string, subnetID int64) *models.NetworkInterface {
	var nic models.NetworkInterface
	_, err := s.db.
		NewSelect().
		Model(&nic).
		Where("mac = ?", mac).
		Where("subnetwork_id = ?", subnetID).
		Relation("Machine").
		Relation("Subnetworks").
		Limit(1).
		Exec(ctx)
	if err != nil {
		s.onError(err)
		return nil
	}
	return &nic
}

func (s *BunStorage) NewEmptyMachine(ctx context.Context) *models.Machine {
	m := new(models.Machine)
	if err := s.db.NewInsert().Model(m).Scan(ctx, m); err != nil {
		s.onError(err)
		return nil
	}
	return m
}

func (s *BunStorage) GetOrCreateSubnetwork(ctx context.Context, cidr string) *models.Subnetwork {
	subnet := new(models.Subnetwork)
	subnet.NetworkCIDR = cidr
	subnet.IPVersion = utils.IPVersionFromCIDR(cidr)
	err := s.db.
		NewInsert().
		Model(subnet).
		On("CONFLICT (network_cidr) DO UPDATE").
		Set("updated_at = CURRENT_TIMESTAMP").
		Scan(ctx, subnet)
	if err != nil {
		s.onError(err)
		return nil
	}
	return subnet
}

func (s *BunStorage) GetAllIPv4Networks(ctx context.Context) []models.Subnetwork {
	subs := make([]models.Subnetwork, 0)
	err := s.db.NewSelect().
		Model((*models.Subnetwork)(nil)).
		Where("ip_version = ?", 4).
		Scan(ctx, &subs)
	if err != nil {
		s.onError(err)
		return nil
	}
	return subs
}
