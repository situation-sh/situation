package cmd

import (
	"context"
	"fmt"

	"github.com/situation-sh/situation/agent/config"
	"github.com/urfave/cli/v3"
)

var extended = false

var versionCmd = cli.Command{
	Name:   "version",
	Usage:  "Print the version of the agent",
	Action: versionAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "extended",
			Aliases:     []string{"e"},
			Required:    false,
			Destination: &extended,
			Usage:       "Print version and commit hash",
		},
	},
}

func versionAction(ctx context.Context, cmd *cli.Command) error {
	if extended {
		fmt.Printf("Version: %s\nCommit: %s\n", config.Version, config.Commit)
	} else {
		fmt.Println(config.Version)
	}
	return nil
}
