package cmd

import (
	"fmt"

	"github.com/situation-sh/situation/config"
	"github.com/urfave/cli/v2"
)

var idCmd = cli.Command{
	Name:   "id",
	Usage:  "Print the identifier of the agent",
	Action: runIDCmd,
}

func runIDCmd(c *cli.Context) error {
	fmt.Println(config.GetAgent())
	return nil
}

func init() {
	app.Commands = append(app.Commands, &idCmd)
}
