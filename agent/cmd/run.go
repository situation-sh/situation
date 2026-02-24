package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/asiffer/puzzle"
	"github.com/getsentry/sentry-go"
	sentrylogrus "github.com/getsentry/sentry-go/logrus"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v3"

	"github.com/situation-sh/situation/agent/config"
	"github.com/situation-sh/situation/pkg/modules"
	"github.com/situation-sh/situation/pkg/store"
	"github.com/situation-sh/situation/pkg/tui"
)

var (
	ignoreMissingDeps bool   = false
	db                string = ":memory:"
	sentryDSN         string = ""
	failfast          bool   = false
	explore           bool   = false
	noMigrate         bool   = false
)

var runCmd = cli.Command{
	Name:   "run",
	Usage:  "Run the agent (default)",
	Action: runAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "explore",
			Destination: &explore,
			Usage:       "Run the explorer after the run",
		},
		&cli.BoolFlag{
			Name:        "no-migrate",
			Destination: &noMigrate,
			Usage:       "Skip database migrations",
		},
	},
}

func init() {
	populateConfig()
	runCmd.Flags = append(runCmd.Flags, generateFlags()...)
	// we first read env before parsing cli parameters
	config.ReadEnv()
}

func disableFlagName(name string) string {
	return fmt.Sprintf("no-module-%s", name)
}

func moduleEnvName(name string) string {
	e := strings.TrimSpace(name)
	e = strings.ReplaceAll(e, "-", "_")
	return strings.ToUpper(e)
}

func defineDB() {
	config.DefineVar(
		"db",
		&db,
		puzzle.WithDescription("Database DSN (e.g. file path for SQLite or connection string for postgres)"),
		puzzle.WithEnvName("SITUATION_DB"),
	)
}

// populateConfig adds configuration variables from modules
// These conf variables will be exported as CLI flags
func populateConfig() {
	config.DefineVar(
		"ignore-missing-deps",
		&ignoreMissingDeps,
		puzzle.WithDescription("Force modules execution even if some required modules are disabled"),
	)
	defineDB()
	config.DefineVar(
		"sentry",
		&sentryDSN,
		puzzle.WithDescription("Sentry DSN for tracing"),
		puzzle.WithEnvName("SENTRY_DSN"),
	)
	config.DefineVar(
		"fail-fast",
		&failfast,
		puzzle.WithDescription("Return directly when a module fails"),
	)

	// config from modules
	modules.Walk(func(name string, mod modules.Module) {
		// add specific config to flags
		if configurableMod, ok := mod.(config.Configurable); ok {
			config.Bind(configurableMod)
		}
		// enable/disable module
		config.Define(
			disableFlagName(name),
			false,
			puzzle.WithDescription(fmt.Sprintf("Disable module %s", name)),
			puzzle.WithEnvName(fmt.Sprintf("NO_MODULE_%s", moduleEnvName(name))),
		)
	})
}

func generateFlags() []cli.Flag {
	flags, err := config.Urfave3()
	if err != nil {
		panic(err)
	}
	sort.Sort(cli.FlagsByName(flags))
	return flags
}

func runAction(ctx context.Context, cmd *cli.Command) error {
	var loggerInterface logrus.FieldLogger = logger
	// fmt.Println("report caller:", logger.ReportCaller)
	// scheduler opts
	opts := make([]modules.SchedulerOptions, 0)

	// sentry integration
	if sentryDSN != "" {
		if err := initSentry(sentryDSN); err != nil {
			return fmt.Errorf("failed to init sentry: %v", err)
		}
		defer sentry.Flush(2 * time.Second)

		// Get the Sentry client from the current hub
		hub := sentry.CurrentHub()
		client := hub.Client()
		if client == nil {
			return fmt.Errorf("sentry client is nil")
		}

		hook := sentrylogrus.NewLogHookFromClient([]logrus.Level{
			// choose what levels are forwarded to Sentry
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.ErrorLevel,
			logrus.FatalLevel,
			logrus.PanicLevel,
		}, client)
		hook.SetHubProvider(func() *sentry.Hub {
			return hub
		})

		// sentry transaction
		tx := sentry.StartTransaction(ctx, "situation.run")
		defer tx.Finish()

		// add scheduler option
		sv := newSentrySupervisor(tx)
		opts = append(opts, modules.WithSupervisor(sv))
		// tx.Status
		// transaction context
		ctx = tx.Context()

		// update the logger
		logger.AddHook(hook)
		loggerInterface = logger.WithContext(ctx)
	}

	storage, err := store.NewStorage(db, config.AgentString(), func(err error) {
		logger.WithField("on", "storage").Warn(err)
	})
	if err != nil {
		logger.Errorf("Failed to create storage: %v", err)
		return fmt.Errorf("failed to create storage: %v", err)
	}

	if !noMigrate {
		logger.WithField("on", "storage").Info("Migrating")
		if err := storage.Migrate(ctx); err != nil {
			logger.Errorf("Failed to migrate: %v", err)
			return fmt.Errorf("failed to migrate: %v", err)
		}
	}

	newCtx := modules.SituationContext(ctx, config.AgentString(), storage, loggerInterface)

	// scheduler opts
	opts = append(opts,
		modules.WithLogger(loggerInterface),
	)
	if ignoreMissingDeps {
		opts = append(opts, modules.IgnoreMissingDeps())
	}
	if failfast {
		opts = append(opts, modules.FailFast())
	}

	// filter modules
	mods := make([]modules.Module, 0)
	modules.Walk(func(name string, m modules.Module) {
		disabled, err := config.Get[bool](disableFlagName(name))
		if err != nil {
			panic(err)
		}
		if !disabled {
			mods = append(mods, m)
		}
	})

	// run the scheduler
	scheduler := modules.NewScheduler(mods, opts...)
	if err := scheduler.Run(newCtx); err != nil {
		return err
	}

	if explore {
		// run the TUI
		return tui.NewRootModel(ctx, storage).Run()
	}

	return nil
}
