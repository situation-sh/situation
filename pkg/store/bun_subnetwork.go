package store

import (
	"context"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/utils"
)

// GetOrCreateSubnetwork returns a subnetwork by its CIDR or creates it if it doesn't exist.
func (s *BunStorage) GetOrCreateSubnetwork(ctx context.Context, cidr string) *models.Subnetwork {
	subnet := new(models.Subnetwork)
	subnet.NetworkCIDR = cidr
	subnet.IPVersion = utils.IPVersionFromCIDR(cidr)
	err := s.db.
		NewInsert().
		Model(subnet).
		On("CONFLICT (network_cidr) DO UPDATE").
		Set("updated_at = CURRENT_TIMESTAMP").
		Scan(ctx, subnet)
	if err != nil {
		s.onError(err)
		return nil
	}
	return subnet
}

// GetAllIPv4Networks returns all IPv4 subnetworks.
func (s *BunStorage) GetAllIPv4Networks(ctx context.Context) []models.Subnetwork {
	subs := make([]models.Subnetwork, 0)
	err := s.db.NewSelect().
		Model((*models.Subnetwork)(nil)).
		Where("ip_version = ?", 4).
		Scan(ctx, &subs)
	if err != nil {
		s.onError(err)
		return nil
	}
	return subs
}
