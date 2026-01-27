package store

import (
	"context"

	"github.com/situation-sh/situation/pkg/models"
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
