package store

import (
	"context"
	"fmt"

	"github.com/situation-sh/situation/pkg/models"
)

func (s *BunStorage) GetLocalUsers(ctx context.Context) ([]models.User, error) {
	machineID := s.GetHostID(ctx)
	if machineID <= 0 {
		return nil, fmt.Errorf("cannot retrieve host ID")
	}

	users := make([]models.User, 0)
	err := s.db.NewSelect().
		Model((*models.User)(nil)).
		Where("machine_id = ?", machineID).
		Scan(ctx, &users)
	return users, err
}
