package store

import (
	"context"
	"slices"
	"sort"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
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

	if s.dialect == dialect.PG {
		return s.postgresInsertPackages(ctx, uniquePkgs)
	}

	// fallback to basic bulk insert
	_, err := s.db.
		NewInsert().
		Model(&uniquePkgs).
		On("CONFLICT (name, version, machine_id) DO UPDATE").
		Set("updated_at = CURRENT_TIMESTAMP").
		Exec(ctx)
	return err
}

// postgresInsertPackages inserts packages using PostgreSQL's optimized bulk insert strategy.
// It uses a temporary staging table and a single upsert query for maximum performance.
func (s *BunStorage) postgresInsertPackages(ctx context.Context, pkgs []*models.Package) error {
	if len(pkgs) == 0 {
		return nil
	}

	// remove duplicates
	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].Name < pkgs[j].Name
	})
	uniquePkgs := slices.CompactFunc(pkgs, func(p1, p2 *models.Package) bool {
		return p1.Name == p2.Name && p1.Version == p2.Version && p1.MachineID == p2.MachineID
	})

	return s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Create temp table that mirrors packages structure
		_, err := tx.ExecContext(ctx, `
			CREATE TEMP TABLE _packages_staging (LIKE packages INCLUDING DEFAULTS)
			ON COMMIT DROP
		`)
		if err != nil {
			return err
		}

		// Bulk insert into staging table (no conflict handling = fast)
		_, err = tx.NewInsert().
			Model(&uniquePkgs).
			ModelTableExpr("_packages_staging").
			Exec(ctx)
		if err != nil {
			return err
		}

		// Upsert from staging to real table
		// Note: Bun doesn't support INSERT...SELECT with ON CONFLICT via fluent API
		// Using SELECT * avoids hardcoding column names in the INSERT part
		_, err = tx.NewRaw(`
			INSERT INTO packages
			SELECT * FROM _packages_staging
			ON CONFLICT (name, version, machine_id) DO UPDATE SET
				updated_at = EXCLUDED.updated_at,
				vendor = EXCLUDED.vendor,
				manager = EXCLUDED.manager,
				install_time_unix = EXCLUDED.install_time_unix,
				files = EXCLUDED.files
		`).Exec(ctx)
		return err
	})
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
