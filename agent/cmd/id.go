package cmd

import (
	"context"
	"fmt"

	"github.com/situation-sh/situation/agent/config"
	"github.com/urfave/cli/v3"
)

var idCmd = cli.Command{
	Name:    "id",
	Aliases: []string{"agent"},
	Usage:   "Print the identifier of the agent",
	Action:  idAction,
}

func idAction(ctx context.Context, cmd *cli.Command) error {
	fmt.Println(config.AgentString())
	return nil
}
