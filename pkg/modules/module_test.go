package modules

import (
	"fmt"
	"testing"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/store"
	"github.com/uptrace/bun"
)

func NewTestingBunStorage(t *testing.T) *store.BunStorage {
	storage, err := store.NewSQLiteBunStorage(
		":memory:",
		"test-agent",
		func(err error) { t.Error(err) },
	)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	return storage
}

func TestStandardProtocolQuery(t *testing.T) {
	storage := NewTestingBunStorage(t)
	str := storage.DB().
		NewUpdate().
		Model((*models.ApplicationEndpoint)(nil)).
		Where("protocol = ?", "tcp").
		Where("application_protocols IS NULL").
		Where("port IN (?)", bun.In(stdPorts())).
		SetColumn("application_protocols", sqlCase(storage)).
		String()
	fmt.Println(str)
}
