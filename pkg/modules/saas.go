package modules

import (
	"context"
	"fmt"

	"github.com/asiffer/puzzle"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/modules/saas"
)

func init() {
	registerModule(&SaaSModule{MaxEndpoints: 50})
}

// Module definition ---------------------------------------------------------

// SaaSModule identifies SaaS applications based on endpoints data
type SaaSModule struct {
	BaseModule

	MaxEndpoints int
}

// Bind binds configuration options for the SaaS module
// -> see config.Configurable interface
func (m *SaaSModule) Bind(config *puzzle.Config) error {
	if err := setDefault(config, m, "max_endpoints", &m.MaxEndpoints, "Maximum number of endpoints to process"); err != nil {
		return err
	}
	return nil
}

func (m *SaaSModule) Name() string {
	return "saas"
}

func (m *SaaSModule) Dependencies() []string {
	return []string{"tls", "ja4"}
}

func (m *SaaSModule) Run(ctx context.Context) error {
	storage := getStorage(ctx)
	logger := getLogger(ctx, m)
	// SaaS module implementation goes here

	endpoints := make([]*models.ApplicationEndpoint, 0, m.MaxEndpoints)

	err := storage.DB().
		NewSelect().
		Model(&endpoints).
		Where("saas IS NULL").
		Limit(m.MaxEndpoints).
		Scan(ctx)
	if err != nil {
		return fmt.Errorf("fail to retrieve endpoints: %w", err)
	}

	if len(endpoints) == 0 {
		logger.Info("No endpoints to process for SaaS detection")
		return nil
	}

	toUpdate := make([]*models.ApplicationEndpoint, 0)

	for _, endpoint := range endpoints {
		detected, saasName, err := saas.Detect(endpoint)
		if err != nil {
			logger.WithError(err).
				WithField("ip", endpoint.Addr).
				WithField("port", endpoint.Port).
				Warn("SaaS detection error")
			continue
		}
		if detected {
			toUpdate = append(toUpdate, endpoint)
			logger.WithField("ip", endpoint.Addr).
				WithField("port", endpoint.Port).
				WithField("saas", saasName).
				Info("SaaS application detected")
		} else {
			logger.WithField("ip", endpoint.Addr).
				WithField("port", endpoint.Port).
				Debug("No SaaS application detected")
		}
	}

	if len(toUpdate) > 0 {
		_, err = storage.DB().
			NewUpdate().
			Model(&toUpdate).
			Column("saas").
			Bulk().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("fail to update SaaS info for endpoints: %w", err)
		}
		logger.WithField("endpoints", len(toUpdate)).Info("SaaS info updated for endpoints")
	} else {
		logger.Info("No SaaS info to update for endpoints")
	}

	return nil
}
