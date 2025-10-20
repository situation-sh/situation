package cmd

import (
	"context"
	"fmt"

	"github.com/asiffer/puzzle/jsonfile"
	"github.com/urfave/cli/v3"
)

var defaultsCmd = cli.Command{
	Name:    "defaults",
	Aliases: []string{"def"},
	Usage:   "Print the default config",
	Action:  runDefaultsCmd,
}

func runDefaultsCmd(ctx context.Context, cmd *cli.Command) error {
	bytes, err := jsonfile.ToJSON(config)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bytes))
	return nil
}
