package store

import (
	"context"
	"testing"
)

func TestMigrate(t *testing.T) {
	storage, err := NewSQLiteBunStorage(":memory:", "test-agent", func(err error) {
		t.Errorf("Storage error: %v", err)
	})
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	if err := storage.Migrate(context.Background()); err != nil {
		t.Fatalf("failed to migrate storage: %v", err)
	}

}
