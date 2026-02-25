package cmd

import (
	"context"
	"fmt"
	"net/mail"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/agent/config"
	"github.com/situation-sh/situation/pkg/utils"

	"github.com/urfave/cli/v3"
)

var logger = utils.NewLogger()

var logLevel uint

var app = &cli.Command{
	Name:    "situation-internal",
	Usage:   "Internal command to manage the repository",
	Version: config.Version,
	Authors: []any{mail.Address{Name: "Alban Siffer", Address: "alban@situation.sh"}},
	Flags: []cli.Flag{
		&cli.UintFlag{
			Name:        "log-level",
			Usage:       "Log level (0: Panic, 1: Fatal, 2: Error, 3: Warn, 4: Info, 5: Debug)",
			Destination: &logLevel,
			Value:       4,
			Aliases:     []string{"l"},
			Local:       false,
		},
	},
	Commands: []*cli.Command{
		&makeMigrationsCmd,
		&modulesDocCmd,
		&dbDocCmd,
	},
	Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
		level := logrus.Level(logLevel)
		if level == logrus.DebugLevel {
			logger.SetReportCaller(true)
		}
		logger.SetLevel(logrus.Level(logLevel))
		return ctx, nil
	},
}

func rootDir() string {
	_, file, _, _ := runtime.Caller(0)
	startDir := path.Dir(file)
	file, err := utils.FindFileUpward(startDir, "go.mod")
	if err != nil {
		panic(fmt.Sprintf("go.mod file not found (%v)", err))
	}
	return path.Dir(file)

}

// Execute executes the root command.
func Execute(ctx context.Context, args []string) error {
	return app.Run(ctx, args)
}
