package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
)

// Cache holds cached values for performance optimization.
type Cache struct {
	HostID int64 // ID of the host machine in the db
}

// BunStorage is the main storage implementation using Bun ORM.
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

// NewStorage creates a new BunStorage instance, auto-detecting the database type.
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

// NewSQLiteBunStorage creates a new BunStorage instance using SQLite.
func NewSQLiteBunStorage(dataSourceName string, agent string, onError func(error)) (*BunStorage, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, dataSourceName)
	if err != nil {
		return nil, err
	}
	if strings.Contains(dataSourceName, ":memory:") {
		// Prevent connection closure from destroying in-memory database
		// see https://bun.uptrace.dev/guide/drivers.html#important-in-memory-database-configuration
		sqldb.SetMaxIdleConns(1000) // Keep connections alive
		sqldb.SetConnMaxLifetime(0) // No connection expiry
		sqldb.SetMaxOpenConns(1)    // Single connection for consistency

	}
	db := bun.NewDB(sqldb, sqlitedialect.New())
	return newStorage(db, agent, onError), nil
}

// NewPostgresBunStorage creates a new BunStorage instance using PostgreSQL.
func NewPostgresBunStorage(dataSourceName string, agent string, onError func(error)) (*BunStorage, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dataSourceName)))
	db := bun.NewDB(sqldb, pgdialect.New())
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return newStorage(db, agent, onError), nil
}

func newPostgresBunStorageNoPing(dataSourceName string, agent string, onError func(error)) (*BunStorage, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dataSourceName)))
	db := bun.NewDB(sqldb, pgdialect.New())
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

// DB returns the underlying bun.DB instance.
func (s *BunStorage) DB() *bun.DB {
	return s.db
}

// ANY generates a dialect-specific ANY expression for the given attribute and value.
// It should be used in WHERE clauses to check if a value exists in a JSON array column.
func (s *BunStorage) ANY(attr string) string {
	switch s.db.Dialect().Name() {
	case dialect.SQLite:
		return fmt.Sprintf("EXISTS (SELECT 1 FROM json_each(%s) WHERE value = ?)", attr)
	case dialect.PG:
		return fmt.Sprintf("? = ANY (%s)", attr)
	default:
		return ""
	}
}

// OVERLAP generates a dialect-specific array overlap expression for checking if any value
// from an array matches any value in a JSON array column.
// Use with ARRAY() to format the values: Where(s.OVERLAP("ip"), s.ARRAY(ips))
func (s *BunStorage) OVERLAP(attr string) string {
	switch s.db.Dialect().Name() {
	case dialect.SQLite:
		return fmt.Sprintf("EXISTS (SELECT 1 FROM json_each(%s) AS j, json_each(?) AS v WHERE j.value = v.value)", attr)
	case dialect.PG:
		return fmt.Sprintf("%s && ?::varchar[]", attr)
	default:
		return ""
	}
}

// ARRAY formats a string slice as a dialect-specific array literal for use with OVERLAP.
func (s *BunStorage) ARRAY(values []string) string {
	switch s.db.Dialect().Name() {
	case dialect.SQLite:
		return "[\"" + strings.Join(values, "\",\"") + "\"]"
	case dialect.PG:
		return "{" + strings.Join(values, ",") + "}"
	default:
		return ""
	}
}
