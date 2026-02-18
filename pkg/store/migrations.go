package store

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/migrate"
)

//go:embed migrations/sqlite/*.sql
var sqliteMigrations embed.FS

//go:embed migrations/postgres/*.sql
var postgresMigrations embed.FS

var TrackedModels = []any{
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
	(*models.Flow)(nil),
	(*models.EndpointPolicy)(nil),
}

// GenerateSchema returns SQL CREATE TABLE statements for all tracked models
func (s *BunStorage) GenerateSchema() string {
	var statements []string
	statements = append(statements, "--- up")
	for _, model := range TrackedModels {
		query := s.db.NewCreateTable().Model(model).IfNotExists()
		statements = append(statements, query.String()+";")
	}
	statements = append(statements, "--- down")
	for i := len(TrackedModels) - 1; i >= 0; i-- {
		model := TrackedModels[i]
		// query := s.db.NewCreateTable().Model(model).IfNotExists()
		query := s.db.NewDropTable().Model(model).IfExists()
		statements = append(statements, query.String()+";")
	}
	return strings.Join(statements, "\n\n")
}

func (s *BunStorage) SnapshotMigrations() (string, string) {
	upStatements := make([]string, 0)
	for _, model := range TrackedModels {
		query := s.db.NewCreateTable().Model(model).WithForeignKeys().IfNotExists()
		upStatements = append(upStatements, query.String()+";")
	}
	downStatements := make([]string, 0)
	for i := len(TrackedModels) - 1; i >= 0; i-- {
		model := TrackedModels[i]
		query := s.db.NewDropTable().Model(model).IfExists()
		downStatements = append(downStatements, query.String()+";")
	}
	return strings.Join(upStatements, "\n\n"), strings.Join(downStatements, "\n\n")
}

// Migrate applies migrations using bun's migration system
func (s *BunStorage) Migrate(ctx context.Context) error {
	migrations := migrate.NewMigrations()

	// select the right migrations subdirectory
	var fsys fs.FS
	var err error = nil
	switch s.db.Dialect().Name() {
	case dialect.SQLite:
		fsys, err = fs.Sub(sqliteMigrations, "migrations/sqlite")
	case dialect.PG:
		fsys, err = fs.Sub(postgresMigrations, "migrations/postgres")
	default:
		return fmt.Errorf("unsupported database dialect: %s", s.db.Dialect().Name())
	}

	if err != nil {
		return fmt.Errorf("failed to get migrations subdirectory: %w", err)
	}

	if err := migrations.Discover(fsys); err != nil {
		return fmt.Errorf("failed to discover migrations: %w", err)
	}

	migrator := migrate.NewMigrator(s.db, migrations)
	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("failed to init migrator: %w", err)
	}
	if _, err := migrator.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}
	return nil
}

// CreateTables directly creates all tables without migrations
func (s *BunStorage) CreateTables(ctx context.Context) error {
	for _, model := range TrackedModels {
		_, err := s.db.NewCreateTable().Model(model).IfNotExists().Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create table for model %T: %w", model, err)
		}
		s.db.NewCreateIndex().Model(model).IfNotExists().Exec(ctx)
	}
	return nil
}
