package cmd

import (
	"context"
	"fmt"

	"github.com/situation-sh/situation/config"
	"github.com/urfave/cli/v3"
)

var idCmd = cli.Command{
	Name:   "id",
	Usage:  "Print the identifier of the agent",
	Action: idAction,
}

func idAction(ctx context.Context, cmd *cli.Command) error {
	fmt.Println(config.GetAgent())
	return nil
}
