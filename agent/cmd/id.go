package cmd

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/urfave/cli/v3"
)

// ID is the unique identifier of the agent instance
// It is a 32 hexchars (16 bytes) random string
var ID = [...]byte{
	202, 254, 202, 254, 202, 254, 202, 254,
	202, 254, 202, 254, 202, 254, 202, 254,
}

const defaultIDHexString = "cafecafecafecafecafecafecafecafe"

func getDefaultID() []byte {
	id, err := hex.DecodeString(defaultIDHexString)
	if err != nil {
		panic(err)
	}
	return id
}

var idCmd = cli.Command{
	Name:   "id",
	Usage:  "Print the identifier of the agent",
	Action: idAction,
}

func idAction(ctx context.Context, cmd *cli.Command) error {
	fmt.Println(uuid.UUID(ID).String())
	return nil
}
