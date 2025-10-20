package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// Version of the agent (it is set during compilation)
var Version = "X.X.X"

var versionCmd = cli.Command{
	Name:   "version",
	Usage:  "Print the version of the agent",
	Action: versionAction,
}

func versionAction(ctx context.Context, cmd *cli.Command) error {
	fmt.Println(Version)
	return nil
}
