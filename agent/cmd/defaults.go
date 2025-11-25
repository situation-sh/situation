package cmd

import (
	"context"
	"fmt"

	"github.com/situation-sh/situation/agent/config"
	"github.com/urfave/cli/v3"
)

var defaultsCmd = cli.Command{
	Name:    "defaults",
	Aliases: []string{"def"},
	Usage:   "Print the default config",
	Action:  runDefaultsCmd,
}

func runDefaultsCmd(ctx context.Context, cmd *cli.Command) error {
	bytes, err := config.JSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bytes))
	return nil
}
