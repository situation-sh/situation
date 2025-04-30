// Package cmd is the entrypoint of the agent
package cmd

import (
	"context"
	"net/mail"
	"os"
	"runtime"

	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/config"

	"github.com/urfave/cli/v3"
)

var logLevel uint = 1

var app = &cli.Command{
	Name:    "situation",
	Usage:   "Just run it",
	Version: config.Version,
	Authors: []any{mail.Address{Name: "Alban Siffer", Address: "alban@situation.sh"}},
	Flags: []cli.Flag{
		&cli.UintFlag{
			Name:        "log-level",
			Usage:       "Log level (0: Panic, 1: Fatal, 2: Error, 3: Warn, 4: Info, 5: Debug)",
			Value:       logLevel,
			Destination: &logLevel,
			Aliases:     []string{"l"},
			Local:       false,
		},
	},
	SuggestCommandFunc: func(commands []*cli.Command, provided string) string {
		// this function is the only hook we can use to both run a default command
		// when NOTHING is provided and to trigger a fatal error is a provided command
		// does not exist
		if provided == "" {
			return runCmd.Name
		}
		for _, c := range commands {
			// fmt.Println(c.Name)
			if c.Name == provided {
				return c.Name
			}
			for _, alias := range c.Aliases {
				if alias == provided {
					return c.Name
				}
			}
		}
		logrus.Fatalf("'%s' is not a valid command.\n", provided)
		return ""
	},
	DefaultCommand: "run",
	Suggest:        true,
	Commands: []*cli.Command{
		&runCmd,
		&refreshIDCmd,
		&defaultsCmd,
		&idCmd,
		&schemaCmd,
		&updateCmd,
		&versionCmd,
	},
	Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
		return ctx, initLog()
	},
}

func initLog() error {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	// ensure log level is between 0 and 5
	if logLevel > 5 {
		logLevel = 5
	}

	if output := os.Stderr; runtime.GOOS == "windows" {
		// Colored output of logrus does not work for windows
		// but we can circumvent it with ansi color codes
		// https://github.com/sirupsen/logrus/issues/172
		logrus.SetOutput(ansicolor.NewAnsiColorWriter(output))
	} else {
		logrus.SetOutput(output)
	}

	// DebugLevel by default
	logrus.SetLevel(logrus.Level(logLevel))
	return nil
}

// Execute executes the root command.
func Execute(ctx context.Context, args []string) error {
	return app.Run(ctx, args)
}
