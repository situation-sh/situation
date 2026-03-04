package store

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

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
	db       *bun.DB
	agent    string
	onError  func(error)
	readOnly bool
	cache    Cache
	dialect  dialect.Name
}

func (s *BunStorage) Dialect() dialect.Name {
	return s.dialect
}

type BunStorageOption func(*BunStorage)

func WithAgent(agent string) BunStorageOption {
	return func(s *BunStorage) {
		s.agent = agent
	}
}

func WithErrorHandler(onError func(error)) BunStorageOption {
	return func(s *BunStorage) {
		s.onError = onError
	}
}

func ReadOnly() BunStorageOption {
	return func(s *BunStorage) {
		s.readOnly = true
	}
}

func newStorage(db *bun.DB, opts ...BunStorageOption) *BunStorage {
	// register m2m models
	db.RegisterModel((*models.ApplicationEndpoint)(nil))
	db.RegisterModel((*models.UserApplication)(nil))
	db.RegisterModel((*models.NetworkInterfaceSubnet)(nil))

	storage := BunStorage{
		db:       db,
		agent:    "",
		onError:  func(err error) {},
		cache:    Cache{HostID: -1},
		dialect:  db.Dialect().Name(),
		readOnly: false,
	}

	for _, opt := range opts {
		opt(&storage)
	}

	return &storage
}

// NewStorage creates a new BunStorage instance, auto-detecting the database type.
func NewStorage(dataSourceName string, opts ...BunStorageOption) (*BunStorage, error) {
	// Simple heuristic to choose between SQLite and Postgres
	switch detectDBType(dataSourceName) {
	case "sqlite":
		return NewSQLiteBunStorage(dataSourceName, opts...)
	case "postgres":
		return NewPostgresBunStorage(dataSourceName, opts...)
	default:
		return nil, fmt.Errorf("cannot detect database type with dsn=%s", dataSourceName)
	}
}

func isReadOnly(opts ...BunStorageOption) bool {
	s := &BunStorage{}
	for _, opt := range opts {
		opt(s)
	}
	return s.readOnly
}

func sqliteCheckReadOnly(dataSourceName string, opts ...BunStorageOption) string {
	// dumyy struct
	if !isReadOnly(opts...) {
		return dataSourceName
	}
	if strings.Contains(dataSourceName, ":memory:") {
		return dataSourceName
	}
	// the file:// prefix (uri format) is important in modernc/sqlite to pass
	// uri parameters like "mode=ro"
	if !strings.HasPrefix(dataSourceName, "file://") {
		dataSourceName = "file://" + dataSourceName
	}
	if !strings.Contains(dataSourceName, "mode=ro") {
		if strings.Contains(dataSourceName, "?") {
			dataSourceName += "&mode=ro"
		} else {
			dataSourceName += "?mode=ro"
		}
	}
	return dataSourceName
}

// NewSQLiteBunStorage creates a new BunStorage instance using SQLite.
func NewSQLiteBunStorage(dataSourceName string, opts ...BunStorageOption) (*BunStorage, error) {
	dsn := sqliteCheckReadOnly(dataSourceName, opts...)
	sqldb, err := sql.Open(sqliteshim.ShimName, dsn)
	if err != nil {
		return nil, err
	}
	if strings.Contains(dsn, ":memory:") || strings.Contains(dsn, "mode=memory") {
		// Prevent connection closure from destroying in-memory database
		// see https://bun.uptrace.dev/guide/drivers.html#important-in-memory-database-configuration
		sqldb.SetMaxIdleConns(1000) // Keep connections alive
		sqldb.SetConnMaxLifetime(0) // No connection expiry
		sqldb.SetMaxOpenConns(1)    // Single connection for consistency
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())
	return newStorage(db, opts...), nil
}

// extractClientSSL parses sslcert and sslkey from the DSN and returns
// a modified DSN without them, along with a pgdriver.Option to set up
// TLS connection if both are present.
func extractClientSSL(dsn string) (string, pgdriver.Option, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return dsn, nil, err
	}
	q := u.Query()
	sslCert := q.Get("sslcert")
	q.Del("sslcert")
	sslKey := q.Get("sslkey")
	q.Del("sslkey")

	u.RawQuery = q.Encode()
	out := u.String()

	if sslCert != "" && sslKey != "" {
		cert, err := tls.LoadX509KeyPair(sslCert, sslKey)
		if err != nil {
			return out, nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		opt := pgdriver.Option(func(conf *pgdriver.Config) {
			if conf.TLSConfig == nil {
				conf.TLSConfig = &tls.Config{
					MinVersion: tls.VersionTLS12,
				}
			}
			conf.TLSConfig.Certificates = append(conf.TLSConfig.Certificates, cert)
		})
		return out, opt, nil
	}

	return u.String(), nil, nil
}

// NewPostgresBunStorage creates a new BunStorage instance using PostgreSQL.
func NewPostgresBunStorage(dataSourceName string, opts ...BunStorageOption) (*BunStorage, error) {
	// remove key-value options that pgdriver doesn't support and handle them manually
	dsn, opt, err := extractClientSSL(dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN: %w", err)
	}

	pgOpts := []pgdriver.Option{
		pgdriver.WithDSN(dsn),
		pgdriver.WithDialTimeout(6 * time.Second),
	}

	// pgdriver does not parse sslcert/sslkey from the DSN; handle them manually.
	// This option runs after WithDSN so it can extend the TLS config it already built.
	if opt != nil {
		pgOpts = append(pgOpts, opt)
	}

	// If read-only mode is requested, set the appropriate connection parameter.
	if isReadOnly(opts...) {
		pgOpts = append(pgOpts,
			pgdriver.WithConnParams(map[string]any{
				"default_transaction_read_only": "on",
			}),
		)
	}

	connector := pgdriver.NewConnector(pgOpts...)
	sqldb := sql.OpenDB(connector)
	db := bun.NewDB(sqldb, pgdialect.New())
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return newStorage(db, opts...), nil
}

func NewPostgresBunStorageNoPing(dataSourceName string, opts ...BunStorageOption) (*BunStorage, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dataSourceName)))
	db := bun.NewDB(sqldb, pgdialect.New())
	return newStorage(db, opts...), nil
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
