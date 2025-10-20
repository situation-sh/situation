package cmd

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/asiffer/puzzle"
	"github.com/asiffer/puzzle/urfave3"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"

	"github.com/situation-sh/situation/pkg/backends"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/modules"
	"github.com/situation-sh/situation/pkg/perf"
	"github.com/situation-sh/situation/pkg/store"
	"github.com/situation-sh/situation/pkg/types"
)

// var (
// 	scans       uint          = 1
// 	period      time.Duration = time.Minute
// 	resetPeriod uint          = 2
// )

// cli config
var config = puzzle.NewConfig()

var ignoreMissingDeps bool = false

func init() {
	// config.DefineVar("scans", &scans, puzzle.WithDescription("Number of scans to perform"))
	// config.DefineVar("period", &period, puzzle.WithDescription("Waiting time between two scans"))
	// config.DefineVar("reset", &resetPeriod, puzzle.WithDescription("Number of runs before resetting the internal store"))
	// app.Flags = append(app.Flags, config.Flags()...)
	// sort.Sort(cli.FlagsByName(app.Flags))
	populateConfig()
	runCmd.Flags = append(runCmd.Flags, generateFlags()...)
}

var runCmd = cli.Command{
	Name:   "run",
	Usage:  "Run the agent (default)",
	Action: runAction,
}

func disableFlagName(name string) string {
	return fmt.Sprintf("no-module-%s", name)
}

func enableBackendKey(name string) string {
	return fmt.Sprintf("enable-backend-%s", name)
}

func populateConfig() {
	puzzle.DefineVar(
		config,
		"ignore-missing-deps",
		&ignoreMissingDeps,
		puzzle.WithDescription("Skip modules with missing dependencies"),
	)

	// config from modules
	modules.Walk(func(name string, mod modules.Module) {
		// add specific config to flags
		if configurableMod, ok := mod.(types.Configurable); ok {
			configurableMod.Bind(config)
		}
		// enable/disable module
		puzzle.Define(
			config,
			disableFlagName(name),
			false,
			puzzle.WithDescription(fmt.Sprintf("Disable module %s", name)))
	})

	// config from backends
	backends.Walk(func(name string, backend backends.Backend) {
		// add specific config to flags
		if configurableBackend, ok := backend.(types.Configurable); ok {
			configurableBackend.Bind(config)
		}
		// enable/disable backend
		puzzle.Define(
			config,
			enableBackendKey(name),
			false,
			puzzle.WithDescription(fmt.Sprintf("Enable the %s backend", name)),
			puzzle.WithFlagName(name),
		)
	})
}

func generateFlags() []cli.Flag {
	if flags, err := urfave3.Build(config); err != nil {
		panic(err)
	} else {
		sort.Sort(cli.FlagsByName(flags))
		// fmt.Println("Generated flags:", flags)
		return flags
	}
}

// func loopCondition(n uint, scans uint, period time.Duration) bool {
// 	if n == scans {
// 		return false
// 	}
// 	time.Sleep(period)
// 	return true
// }

func runAction(ctx context.Context, cmd *cli.Command) error {
	begin = time.Now()

	logger := logrus.New()
	logger.Formatter = &ModuleFormatter{}
	storage := store.NewMemoryStore(ID)
	enabledBackends := make([]backends.Backend, 0)

	// init backends
	for backend := range backends.Iterate() {
		enabled, err := puzzle.Get[bool](config, enableBackendKey(backend.Name()))
		if err != nil {
			return err
		}
		if enabled {
			if logProducer := backend.(types.LogProducer); logProducer != nil {
				logProducer.SetLogger(logger.WithField("backend", backend.Name()))
			}
			logger.Infof("Backend %s enabled", backend.Name())
			if err := backend.Init(); err != nil {
				return fmt.Errorf("failed to init backend %s: %w", backend.Name(), err)
			}
			enabledBackends = append(enabledBackends, backend)
		} else {
			logger.Infof("Backend %s disabled", backend.Name())
		}
	}

	// init modules and scheduler
	mods := make([]modules.Module, 0)
	modules.Walk(func(name string, m modules.Module) {
		disabled, err := puzzle.Get[bool](config, disableFlagName(name))
		if err != nil {
			panic(err)
		}
		if !disabled {
			// add logger
			if logProducer := m.(types.LogProducer); logProducer != nil {
				logProducer.SetLogger(logger.WithField("module", name))
			}
			// add storage
			if storable := m.(types.StorageDemander); storable != nil {
				storable.SetStore(storage)
			}
			mods = append(mods, m)
			logger.Infof("Module %s enabled", name)
		} else {
			logger.Infof("Module %s disabled", name)
		}
	})
	scheduler := modules.NewScheduler(
		mods,
		modules.WithLogger(logger.WithField("module", "[scheduler]")),
		modules.IgnoreMissingDeps(ignoreMissingDeps),
	)

	// run
	start := time.Now()
	if err := scheduler.Run(); err != nil {
		return err
	}
	end := time.Now()

	// prepare payload
	perfs := perf.Collect()
	payload := storage.InitPayload()
	payload.Extra = &models.ExtraInfo{
		Agent:     ID,
		Version:   Version,
		Timestamp: end,
		Duration:  end.Sub(start),
		Errors:    modules.BuildModuleErrors(),
		Perfs:     perfs,
	}

	// send to backends
	for _, backend := range enabledBackends {
		if err := backend.Write(payload); err != nil {
			logger.Errorf("Failed to write to backend %s: %v", backend.Name(), err)
		}
	}

	// for n, run := uint(0), true; run; n, run = n+1, loopCondition(n+1, scans, period) {
	// 	// scan
	// 	if err := singleRun(); err != nil {
	// 		return err
	// 	}
	// 	// reset internal store
	// 	if n > 0 && n%resetPeriod == 0 {
	// 		store.Clear()
	// 	}
	// }

	return nil
}

// func singleRun() error {
// 	// run the modules
// 	scheduler := modules.NewScheduler(modules.G)
// 	scheduler
// 	start := time.Now()
// 	if err := modules.RunModules(); err != nil {
// 		return err
// 	}
// 	end := time.Now()

// 	// collect memory performances after the scan
// 	perfs := perf.Collect()

// 	// fill the payload
// 	payload := store.InitPayload()
// 	payload.Extra = &models.ExtraInfo{
// 		Agent:     config.GetAgent(),
// 		Version:   config.Version,
// 		Timestamp: end,
// 		Duration:  end.Sub(start),
// 		Errors:    modules.BuildModuleErrors(),
// 		Perfs:     perfs,
// 	}

// 	return backends.Write(payload)
// }
