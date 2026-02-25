package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/agent/cmd"
)

func main() {
	ctx := context.Background()
	if err := cmd.Execute(ctx, os.Args); err != nil {
		logrus.Fatal(err)
	}
}
