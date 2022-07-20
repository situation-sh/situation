package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/situation-sh/situation/models"
	"github.com/urfave/cli/v2"
)

var schemaCmd = cli.Command{
	Name:   "schema",
	Usage:  "Print the JSON schema of the data exported by this agent",
	Action: runSchemaCmd,
}

func runSchemaCmd(c *cli.Context) error {
	schema := jsonschema.Reflect(&models.Payload{})
	data, _ := json.MarshalIndent(schema, "", "  ")
	fmt.Println(string(data))
	return nil
}

func init() {
	app.Commands = append(app.Commands, &schemaCmd)
}
