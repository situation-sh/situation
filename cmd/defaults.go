package cmd

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/situation-sh/situation/utils"
	"gopkg.in/yaml.v3"
)

var defaultsCmd = cli.Command{
	Name:    "defaults",
	Aliases: []string{"def"},
	Usage:   "Print the default config",
	Action:  runDefaultsCmd,
	Before:  before,
}

func insert(in map[string]interface{}, key string, value interface{}) error {
	subs := strings.SplitN(key, ".", 2)
	if len(subs) == 1 {
		in[key] = value
		return nil
	}

	any, exists := in[subs[0]]
	if exists {
		if m, ok := any.(map[string]interface{}); ok {
			return insert(m, subs[1], value)
		}
		return fmt.Errorf("wrong type for key %s (%T)", subs[0], subs[0])
	}

	in[subs[0]] = make(map[string]interface{})
	// cannot fail
	if m, ok := in[subs[0]].(map[string]interface{}); ok {
		return insert(m, subs[1], value)
	}
	return nil
}

func unflat(in map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	for key, value := range in {
		if err := insert(out, key, value); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func runDefaultsCmd(c *cli.Context) error {
	u, err := unflat(utils.BuiltFlags())
	if err != nil {
		return err
	}

	out, err := yaml.Marshal(u)
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func init() {
	app.Commands = append(app.Commands, &defaultsCmd)
}
