package cmd

import (
	"context"
	"fmt"

	"github.com/situation-sh/situation/config"
	"github.com/urfave/cli/v3"
)

var defaultsCmd = cli.Command{
	Name:    "defaults",
	Aliases: []string{"def"},
	Usage:   "Print the default config",
	Action:  runDefaultsCmd,
}

func runDefaultsCmd(ctx context.Context, cmd *cli.Command) error {
	out := config.JSON()
	fmt.Println(string(out))
	return nil
}
