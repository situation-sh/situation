package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/situation-sh/situation/pkg/store"
	"github.com/urfave/cli/v3"
)

const initialMigrationName = "0001_initial"

var (
	defaultMigrationBaseDir string = defaultMigrationsDir()
	noSqlite                bool   = false
	noPostgres              bool   = false
	overrideInitial         bool   = false
	sqliteMigrationDir      string
	postgresMigrationDir    string
	migrationName           string
)

func defaultMigrationsDir() string {
	return path.Join(rootDir(), "pkg", "store", "migrations")
}

const MakeMigrationsDescription = `This command generates up/down SQL migration files for the database.
It does not generate diffs but rather a snapshot of the current 
state of the database. The latter can then be used to manually 
write the migration you need.`

var makeMigrationsCmd = cli.Command{
	Name:        "makemigrations",
	Usage:       "Generate migration files for the database",
	Description: MakeMigrationsDescription,
	Action:      makeMigrationsAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "no-sqlite",
			Value:       false,
			Destination: &noSqlite,
		},
		&cli.BoolFlag{
			Name:        "no-postgres",
			Value:       false,
			Destination: &noPostgres,
		},
		&cli.StringFlag{
			Name:        "sqlite-migration-dir",
			Value:       path.Join(defaultMigrationBaseDir, "sqlite"),
			Destination: &sqliteMigrationDir,
		},
		&cli.StringFlag{
			Name:        "postgres-migration-dir",
			Value:       path.Join(defaultMigrationBaseDir, "postgres"),
			Destination: &postgresMigrationDir,
		},
		&cli.StringFlag{
			Name:        "migration-name",
			Value:       "9999_UNNAMED",
			Destination: &migrationName,
			Usage:       "Name of the migration file. Must be in the format <NUMBER>_<NAME>",
		},
		&cli.BoolFlag{
			Name:        "override-initial",
			Value:       false,
			Destination: &overrideInitial,
			Usage:       "Override the initial migration if it already exists. It sets migration name to 0000_initial and override the existing migration files.",
		},
	},
}

func generateMigrations(storage *store.BunStorage, migrationsDir string, migrationName string) error {
	up, down := storage.SnapshotMigrations()
	if err := os.WriteFile(
		path.Join(migrationsDir, fmt.Sprintf("%s.up.sql", migrationName)),
		[]byte(up),
		0644,
	); err != nil {
		return fmt.Errorf("failed to write up migration: %v", err)
	}

	if err := os.WriteFile(
		path.Join(migrationsDir, fmt.Sprintf("%s.down.sql", migrationName)),
		[]byte(down),
		0644,
	); err != nil {
		return fmt.Errorf("failed to write down migration: %v", err)
	}
	return nil
}

func makeMigrationsAction(ctx context.Context, cmd *cli.Command) error {
	if overrideInitial {
		migrationName = initialMigrationName
	}
	if !noSqlite {
		logger.Info("Creating SQLite storage")
		storage, err := store.NewSQLiteBunStorage(":memory:", "test-agent", func(err error) {
			logger.WithError(err).Error("Storage error")
		})
		if err != nil {
			return fmt.Errorf("failed to create storage: %v", err)
		}
		logger.
			WithField("location", sqliteMigrationDir).
			WithField("migration_name", migrationName).
			Info("Creating migration file")
		if err := generateMigrations(storage, sqliteMigrationDir, migrationName); err != nil {
			return fmt.Errorf("failed to generate SQLite migrations: %v", err)
		}
	}
	if !noPostgres {
		logger.Info("Creating Postgres storage")
		storage, err := store.NewPostgresBunStorageNoPing("postgresql://postgres:situation@localhost:5432/situation", "test-agent", func(err error) {
			logger.WithError(err).Error("Storage error")
		})
		if err != nil {
			return fmt.Errorf("failed to create storage: %v", err)
		}
		logger.
			WithField("location", postgresMigrationDir).
			WithField("migration_name", migrationName).
			Info("Creating migration file")
		if err := generateMigrations(storage, postgresMigrationDir, migrationName); err != nil {
			return fmt.Errorf("failed to generate Postgres migrations: %v", err)
		}
	}
	return nil
}
