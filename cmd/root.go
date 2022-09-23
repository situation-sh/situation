// Package cmd is the entrypoint of the agent
package cmd

import (
	"os"
	"runtime"
	"time"

	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/backends"
	"github.com/situation-sh/situation/config"
	"github.com/situation-sh/situation/modules"
	"github.com/situation-sh/situation/utils"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

// contextConfigKey is a type for config keys
// stored in context
type contextConfigKey string

var flags = []cli.Flag{
	&cli.PathFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "Path to configuration file",
	},
}

var app = &cli.App{
	Name:                 "situation",
	Usage:                "Just run it",
	Version:              config.Version,
	Authors:              []*cli.Author{{Name: "Alban Siffer", Email: "alban@situation.sh"}},
	Action:               runRunCmd,
	Before:               before,
	Flags:                make([]cli.Flag, 0),
	EnableBashCompletion: true,
}

func init() {
	// as 'cmd' calls 'modules' and 'backends' the latters
	// are initialized first so DefaultFlags are well filled
	flags = append(flags, rootFlags()...)
	flags = append(flags, modules.DefaultFlags...)
	flags = append(flags, backends.DefaultFlags...)
	app.Flags = append(app.Flags, flags...)
}

func rootFlags() []cli.Flag {
	// utils.BuildFlag is called to centralize the creation of options
	scans := utils.BuildFlag("scans", 1, "Number of scans to perform", []string{"s"})
	period := utils.BuildFlag("period", 2*time.Minute, "Waiting time between two scans", []string{"p"})
	logLevel := utils.BuildFlag("log-level", 0, "Panic: 0, Fatal: 1, Error: 2, Warn: 3, Info: 4, Debug: 5", []string{"l"})
	// logFile := utils.BuildFlag("log-to-file", "", "Redirect logs to file instead of stderr", []string{"f"})
	return []cli.Flag{scans, period, logLevel}
}

func initLog() error {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})

	logLevel, err := config.Get[int]("log-level")
	if err != nil {
		return err
	}
	// ensure log level is between 0 and 5
	if logLevel < 0 {
		logLevel = 0
	} else if logLevel > 5 {
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

func before(c *cli.Context) error {
	fun := altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc("config"))

	if err := fun(c); err != nil {
		return err
	}

	// prepare the config for the modules and the backends
	config.InjectContext(c)

	if err := initLog(); err != nil {
		return err
	}
	return nil
}

// Execute executes the root command.
func Execute() error {
	return app.Run(os.Args)
}
