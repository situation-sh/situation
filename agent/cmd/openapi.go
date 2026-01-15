package cmd

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
)

var openapiCmd = cli.Command{
	Name:   "openapi",
	Usage:  "Output OpenAPI specification",
	Action: openapiAction,
}

func openapiAction(ctx context.Context, cmd *cli.Command) error {
	api, _, _ := initAPI()

	bytes, err := api.OpenAPI().MarshalJSON()
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(bytes)
	return err
}
