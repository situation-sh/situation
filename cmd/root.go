// Package cmd is the entrypoint of the agent
package cmd

import (
	"context"
	"net/mail"
	"os"
	"runtime"
	"sort"

	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/config"

	// "github.com/urfave/cli/altsrc"
	"github.com/urfave/cli/v3"
	// "github.com/urfave/cli/v2/altsrc"
)

// contextConfigKey is a type for config keys
// stored in context
// type contextConfigKey string

// var flags = []cli.Flag{
// 	&cli.PathFlag{
// 		Name:    "config",
// 		Aliases: []string{"c"},
// 		Usage:   "Path to configuration file",
// 	},
// }

//	var app = &cli.App{
//		Name:                 "situation",
//		Usage:                "Just run it",
//		Version:              config.Version,
//		Authors:              []*cli.Author{{Name: "Alban Siffer", Email: "alban@situation.sh"}},
//		Action:               runRunCmd,
//		Before:               before,
//		Flags:                make([]cli.Flag, 0),
//		EnableBashCompletion: true,
//	}

var logLevel uint = 0

var app = &cli.Command{
	Name:    "situation",
	Usage:   "Just run it",
	Version: config.Version,
	Authors: []any{mail.Address{Name: "Alban Siffer", Address: "alban@situation.sh"}},
	Action:  runAction,
	Flags:   make([]cli.Flag, 0),
	Commands: []*cli.Command{
		&refreshIDCmd,
		&defaultsCmd,
		&idCmd,
		&schemaCmd,
	},
}

func init() {
	config.DefineVarWithUsage(
		"log-level",
		&logLevel,
		"Log level (0: Panic, 1: Fatal, 2: Error, 3: Warn, 4: Info, 5: Debug)",
	)
	app.Flags = append(app.Flags, config.Flags()...)
	// sort app.Flags
	sort.Sort(cli.FlagsByName(app.Flags))
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
	if err := initLog(); err != nil {
		return err
	}
	return app.Run(ctx, args)
}
