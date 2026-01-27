package store

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"
)

func TestGenerateSchema(t *testing.T) {
	storage, err := NewSQLiteBunStorage(":memory:", "test-agent", func(err error) {
		t.Errorf("Storage error: %v", err)
	})
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	sql := storage.GenerateSchema()
	fmt.Printf("%s\n", sql)
	// t.Logf("Generated SQL:\n%s", sql)
}

func TestMigrate(t *testing.T) {
	storage, err := NewSQLiteBunStorage(":memory:", "test-agent", func(err error) {
		t.Errorf("Storage error: %v", err)
	})
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	if err := storage.Migrate(context.Background()); err != nil {
		t.Fatalf("failed to migrate tables: %v", err)
	}
}

func generateMigrations(t *testing.T, storage *BunStorage, subdir string) {
	_, file, _, _ := runtime.Caller(0)
	migrationsDir := path.Join(path.Dir(file), "migrations", subdir)

	up, down := storage.snapshotMigrations()
	if err := os.WriteFile(
		path.Join(migrationsDir, "9999_UNNAMED.up.sql"),
		[]byte(up),
		0644,
	); err != nil {
		t.Fatalf("failed to write up migration: %v", err)
	}

	if err := os.WriteFile(
		path.Join(migrationsDir, "9999_UNNAMED.down.sql"),
		[]byte(down),
		0644,
	); err != nil {
		t.Fatalf("failed to write down migration: %v", err)
	}
}

// TestGenerateSQLiteMigrations generates the latest migration SQL files for SQLite
// The files must then be diffed to properly write the migration scripts.
func TestGenerateSQLiteMigrations(t *testing.T) {
	storage, err := NewSQLiteBunStorage(":memory:", "test-agent", func(err error) {
		t.Errorf("Storage error: %v", err)
	})
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	generateMigrations(t, storage, "sqlite")
}

// TestGeneratePostgresMigrations generates the latest migration SQL files for PostgreSQL
// The files must then be diffed to properly write the migration scripts.
func TestGeneratePostgresMigrations(t *testing.T) {
	// init a fake postgres storage (without pinging)
	storage, err := newPostgresBunStorageNoPing("postgres://user:pass@localhost:5432/dbname?sslmode=disable", "test-agent", func(err error) {
		t.Errorf("Storage error: %v", err)
	})
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	generateMigrations(t, storage, "postgres")
}
