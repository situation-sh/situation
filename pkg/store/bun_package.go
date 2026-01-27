package store

import (
	"context"
	"slices"
	"sort"

	"github.com/situation-sh/situation/pkg/models"
)

// InsertPackages inserts a list of packages into the database.
// Duplicates are removed before insertion.
func (s *BunStorage) InsertPackages(ctx context.Context, pkgs []*models.Package) error {
	// remove duplicates
	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].Name < pkgs[j].Name
	})
	uniquePkgs := slices.CompactFunc(pkgs, func(p1, p2 *models.Package) bool {
		return p1.Name == p2.Name && p1.Version == p2.Version && p1.MachineID == p2.MachineID
	})
	_, err := s.db.
		NewInsert().
		Model(&uniquePkgs).
		On("CONFLICT (name, version, machine_id) DO UPDATE").
		Set("updated_at = CURRENT_TIMESTAMP").
		Exec(ctx)
	return err
}

// FindPackageByApplicationName finds a package by the application filename it contains.
func (s *BunStorage) FindPackageByApplicationName(ctx context.Context, hostId int64, filename string) (*models.Package, error) {
	pkg := models.Package{}
	err := s.db.
		NewSelect().
		Model(&pkg).
		Where("machine_id = ?", hostId).
		Where(s.ANY("files"), filename).
		Relation("Applications").
		Scan(ctx)
	if err != nil {
		s.onError(err)
		return nil, err
	}

	return &pkg, nil
}
