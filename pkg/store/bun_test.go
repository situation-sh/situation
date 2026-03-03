package store

import (
	"context"
	"fmt"
	"testing"
)

func TestGenerateSchema(t *testing.T) {
	storage, err := NewSQLiteBunStorage(":memory:",
		WithAgent("test-agent"),
		WithErrorHandler(func(err error) {
			t.Errorf("Storage error: %v", err)
		}),
	)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	sql := storage.GenerateSchema()
	fmt.Printf("%s\n", sql)
	// t.Logf("Generated SQL:\n%s", sql)
}

func TestMigrateSQLite(t *testing.T) {
	storage, err := NewSQLiteBunStorage(":memory:",
		WithAgent("test-agent"),
		WithErrorHandler(func(err error) {
			t.Errorf("Storage error: %v", err)
		}),
	)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	if err := storage.Migrate(context.Background()); err != nil {
		t.Fatalf("failed to migrate tables: %v", err)
	}
}

func TestQLiteCheckReadOnly(t *testing.T) {
	dsn := "/tmp/test.db"
	rodsn := sqliteCheckReadOnly(dsn, ReadOnly())
	if rodsn != "file://"+dsn+"?mode=ro" {
		t.Errorf("Expected read-only DSN to be '%s?mode=ro', got '%s'", dsn, rodsn)
	}

	dsn = "/tmp/test.db?cache=shared"
	rodsn = sqliteCheckReadOnly(dsn, ReadOnly())
	if rodsn != "file://"+dsn+"&mode=ro" {
		t.Errorf("Expected read-only DSN to be '%s&mode=ro', got '%s'", dsn, rodsn)
	}

	dsn = "test.db?mode=ro"
	rodsn = sqliteCheckReadOnly(dsn, ReadOnly())
	if rodsn != "file://"+dsn {
		t.Errorf("Expected read-only DSN to be '%s&mode=ro', got '%s'", dsn, rodsn)
	}
}
