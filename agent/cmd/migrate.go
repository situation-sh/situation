package cmd

import (
	"context"
	"fmt"

	"github.com/situation-sh/situation/agent/config"
	"github.com/situation-sh/situation/pkg/store"
	"github.com/urfave/cli/v3"
)

var migrateCmd = cli.Command{
	Name:   "migrate",
	Usage:  "Migrate the database schema",
	Action: migrateAction,
}

func init() {
	defineDB()
	flags, err := config.SomeFlags("db")
	if err != nil {
		panic(err)
	}
	migrateCmd.Flags = append(migrateCmd.Flags, flags...)
}

func migrateAction(ctx context.Context, cmd *cli.Command) error {
	storage, err := store.NewStorage(db, config.AgentString(), func(err error) {
		logger.WithField("on", "storage").Warn(err)
	})
	if err != nil {
		logger.Errorf("Failed to create storage: %v", err)
		return fmt.Errorf("failed to create storage: %v", err)
	}

	logger.WithField("on", "storage").Info("Migrating")
	if err := storage.Migrate(ctx); err != nil {
		logger.Errorf("Failed to migrate: %v", err)
		return fmt.Errorf("failed to migrate: %v", err)
	}
	return nil
}
