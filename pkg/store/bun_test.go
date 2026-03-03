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
