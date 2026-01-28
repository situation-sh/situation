// Package cmd is the entrypoint of the agent
package cmd

import (
	"context"
	"net/mail"

	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/agent/config"

	"github.com/urfave/cli/v3"
)

var logLevel uint

var app = &cli.Command{
	Name:    "situation",
	Usage:   "Just run it",
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
	// Action: runAction,
	DefaultCommand: runCmd.Name,
	Commands: []*cli.Command{
		&runCmd,
		&refreshIDCmd,
		&defaultsCmd,
		&idCmd,
		&updateCmd,
		&versionCmd,
		&taskCmd,
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

// Execute executes the root command.
func Execute(ctx context.Context, args []string) error {
	return app.Run(ctx, args)
}
