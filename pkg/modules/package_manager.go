package modules

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/store"
	"github.com/situation-sh/situation/pkg/utils"
)

type AbstractPackageManager struct {
	families []string
	hostID   int64
	ctx      context.Context
	logger   logrus.FieldLogger
	storage  *store.BunStorage
}

func NewAbstractPackageManager(ctx context.Context, families []string, logger logrus.FieldLogger, storage *store.BunStorage) (*AbstractPackageManager, error) {
	host := storage.GetOrCreateHost(ctx)
	if host == nil || host.ID == 0 {
		return nil, fmt.Errorf("no host found in storage")
	}

	if host.DistributionFamily != "" && !utils.Includes(families, host.DistributionFamily) {
		logger.
			WithField("distribution_family", host.DistributionFamily).
			Warn("Module skipped for this distribution")
		return nil, &notApplicableError{msg: "distribution family not supported"}
	}

	return &AbstractPackageManager{
		families: families,
		hostID:   host.ID,
		ctx:      ctx,
		logger:   logger,
		storage:  storage,
	}, nil
}

func (a *AbstractPackageManager) Run(generator <-chan *models.Package) error {
	ctx := a.ctx
	logger := a.logger
	storage := a.storage

	appsToUpdate := make([]*models.Application, 0)
	fileAppMap, err := storage.BuildFileAppMap(ctx)
	if err != nil {
		// only warn here
		logger.Warn(err)
	}

	packages := make([]*models.Package, 0)
	for p := range generator {
		p.MachineID = a.hostID
		// log the package found
		logger.WithField("name", p.Name).
			WithField("version", p.Version).
			WithField("install", time.Unix(p.InstallTimeUnix, 0)).
			WithField("files", len(p.Files)).
			Debug("Package found")

		// look at the files and update apps accordingly
		for _, f := range p.Files {
			if _, exists := fileAppMap[f]; exists {
				for _, app := range fileAppMap[f] {
					// link package to app (not the ID yet)
					app.Package = p
					appsToUpdate = append(appsToUpdate, app)
					logger.
						WithField("app", app.Name).
						WithField("file", f).
						Debug("Linking application to package")
				}
			}
		}
		packages = append(packages, p)
	}

	err = storage.InsertPackages(ctx, packages)
	if err != nil {
		return err
	} else {
		logger.WithField("packages", len(packages)).Info("Packages found")
	}

	if len(appsToUpdate) > 0 {
		// update apps ID
		for _, app := range appsToUpdate {
			if app.Package != nil && app.Package.ID != 0 {
				app.PackageID = app.Package.ID
			}
		}

		_, err = storage.DB().
			NewUpdate().
			Model(&appsToUpdate).
			Column("package_id").
			Bulk().
			Exec(ctx)

		logger.
			WithField("apps", len(appsToUpdate)).
			Info("Applications linked to packages")

		return err
	} else {
		logger.Warn("no applications to link to packages")
		return nil
	}
}
