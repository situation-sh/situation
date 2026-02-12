package cmd

import (
	"context"
	"fmt"

	"github.com/situation-sh/situation/agent/config"
	"github.com/situation-sh/situation/pkg/store"
	"github.com/situation-sh/situation/pkg/tui"
	"github.com/urfave/cli/v3"
)

var exploreCmd = cli.Command{
	Name:   "explore",
	Usage:  "View collected data",
	Action: exploreAction,
}

func init() {
	defineDB()
	flags, err := config.SomeFlags("db")
	if err != nil {
		panic(err)
	}
	exploreCmd.Flags = append(exploreCmd.Flags, flags...)
}

func exploreAction(ctx context.Context, cmd *cli.Command) error {
	storage, err := store.NewStorage(db, config.AgentString(), func(err error) {
		// TODO
	})
	if err != nil {
		return fmt.Errorf("failed to create storage: %v", err)
	}

	return tui.NewRootModel(ctx, storage).Run()
}
