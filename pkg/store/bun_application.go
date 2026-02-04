package store

import (
	"context"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/uptrace/bun/dialect"
)

// FindApplicationByName finds an application by its name on a given host.
func (s *BunStorage) FindApplicationByName(ctx context.Context, hostId int64, name string) (*models.Application, error) {
	app := models.Application{}
	err := s.db.
		NewSelect().
		Model(&app).
		Where("machine_id = ? AND name = ?", hostId, name).
		Relation("Users").
		Relation("NetworkInterfaces").
		Relation("Package").
		Relation("Machine").
		Relation("Endpoints").
		Scan(ctx)
	if err != nil {
		s.onError(err)
		return nil, err
	}

	return &app, nil
}

func (s *BunStorage) WithoutJA4() string {
	switch s.dialect {
	case dialect.SQLite:
		return "json_extract(fingerprints, '$.ja4') IS NULL"
	case dialect.PG:
		return "fingerprints::jsonb -> 'ja4' IS NULL"
	default:
		return ""
	}
}
