package cmd

import (
	"context"
	"fmt"

	"github.com/situation-sh/situation/config"
	"github.com/urfave/cli/v3"
)

var versionCmd = cli.Command{
	Name:   "version",
	Usage:  "Print the version of the agent",
	Action: versionAction,
}

func versionAction(ctx context.Context, cmd *cli.Command) error {
	fmt.Println(config.Version)
	return nil
}
