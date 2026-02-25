package cmd

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/minio/selfupdate"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/agent/config"
	"github.com/situation-sh/situation/pkg/utils"
	"github.com/urfave/cli/v3"
)

var refreshIDCmd = cli.Command{
	Name:  "refresh-id",
	Usage: "Regenerate the internal ID of the agent",
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name:  "id",
			Value: hex.EncodeToString(utils.RandBytes(16)),
		},
	},
	Action: refreshIDAction,
}

func refreshIDAction(ctx context.Context, cmd *cli.Command) error {
	newId, err := hex.DecodeString(cmd.StringArg("id"))
	if err != nil {
		return err
	}
	if len(newId) < 16 {
		return fmt.Errorf("the provided id must have 16 bytes")
	}
	logrus.
		WithField("bytes", newId).
		WithField("hex", cmd.StringArg("id")).
		Debug("New ID to set")

	// get the path to the current executable
	binaryFile, err := os.Executable()
	if err != nil {
		return err
	}
	binaryFile = filepath.Clean(binaryFile)

	// see https://pkg.go.dev/os#Executable
	binaryFile, err = filepath.EvalSymlinks(binaryFile)
	if err != nil {
		return err
	}

	// file exist
	if _, err := os.Stat(binaryFile); err != nil {
		return err
	}

	raw, err := os.ReadFile(binaryFile) //#nosec (https://github.com/securego/gosec/issues/821)
	if err != nil {
		return err
	}

	// set a new random ID
	toWrite := bytes.Replace(raw, config.Agent[:16], newId[:16], 1)
	// turn toWrite into is.Reader
	if err := selfupdate.Apply(bytes.NewReader(toWrite), selfupdate.Options{}); err != nil {
		return err
	}

	logrus.
		WithField("bytes", newId).
		WithField("hex", cmd.StringArg("id")).
		Info("ID has been refreshed")
	return nil
}
