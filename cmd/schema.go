package cmd

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/situation-sh/situation/models"
	"github.com/urfave/cli/v2"
)

var schemaCmd = cli.Command{
	Name:   "schema",
	Usage:  "Print the JSON schema of the data exported by this agent",
	Action: runSchemaCmd,
}

// mapper is a custom jsonschema mapper
func mapper(t reflect.Type) *jsonschema.Schema {
	switch t.String() {
	case "models.IP":
		return &jsonschema.Schema{
			AnyOf:    []*jsonschema.Schema{{Type: "string", Format: "ipv4"}, {Type: "string", Format: "ipv6"}},
			Title:    "IPv4 or IPv6 address",
			Examples: []interface{}{"192.168.10.103", "0.0.0.0", "::", "fe80::c1b2:a320:f799:10e0"},
		}
	case "uuid.UUID":
		return &jsonschema.Schema{
			Type:     "string",
			Format:   "uuid",
			Title:    "Universally unique identifier",
			Examples: []interface{}{"123e4567-e89b-12d3-a456-426652340000", "df5591bf-39f6-41bb-b7f6-ba57f2d64309"},
		}
	case "net.HardwareAddr":
		return &jsonschema.Schema{
			Type:     "string",
			Title:    "MAC address",
			Pattern:  "^([0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}$",
			Examples: []interface{}{"5E:FF:56:A2:AF:15", "FF:FF:FF:FF:FF:FF"},
		}
	default:
		return nil
	}
}

func isTesting(c *cli.Context) bool {
	switch b := c.Context.Value("testing").(type) {
	case bool:
		return b
	default:
		return false
	}
}

func runSchemaCmd(c *cli.Context) error {
	reflector := jsonschema.Reflector{Mapper: mapper}

	if !isTesting(c) {
		if err := reflector.AddGoComments("github.com/situation-sh/situation/models", "models"); err != nil {

			return err
		}
	}

	schema := reflector.Reflect(&models.Payload{})
	// we manually add these definition since the User prameter of the Application is an interface{}
	schema.Definitions["LinuxID"] = reflector.Reflect(&models.LinuxID{}).Definitions["LinuxID"]
	schema.Definitions["LinuxUser"] = reflector.Reflect(&models.LinuxUser{}).Definitions["LinuxUser"]
	schema.Definitions["WindowsUser"] = reflector.Reflect(&models.WindowsUser{}).Definitions["WindowsUser"]

	data, _ := json.MarshalIndent(schema, "", "  ")
	// data = data[0:]
	fmt.Println(string(data))
	return nil
}

func init() {
	app.Commands = append(app.Commands, &schemaCmd)
}
