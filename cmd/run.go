package cmd

import (
	"context"
	"sort"
	"time"

	"github.com/asiffer/puzzle"
	"github.com/urfave/cli/v3"

	"github.com/situation-sh/situation/backends"
	"github.com/situation-sh/situation/config"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/modules"
	"github.com/situation-sh/situation/perf"
	"github.com/situation-sh/situation/store"
)

var (
	scans       uint          = 1
	period      time.Duration = time.Minute
	resetPeriod uint          = 2
)

func init() {
	config.DefineVar("scans", &scans, puzzle.WithDescription("Number of scans to perform"))
	config.DefineVar("period", &period, puzzle.WithDescription("Waiting time between two scans"))
	config.DefineVar("reset", &resetPeriod, puzzle.WithDescription("Number of runs before resetting the internal store"))
	app.Flags = append(app.Flags, config.Flags()...)
	sort.Sort(cli.FlagsByName(app.Flags))
}

var runCmd = cli.Command{
	Name:   "run",
	Usage:  "Run the agent (default)",
	Action: runAction,
}

func loopCondition(n uint, scans uint, period time.Duration) bool {
	if n == scans {
		return false
	}
	time.Sleep(period)
	return true
}

func runAction(ctx context.Context, cmd *cli.Command) error {
	if err := backends.Init(); err != nil {
		return err
	}
	defer backends.Close()

	for n, run := uint(0), true; run; n, run = n+1, loopCondition(n+1, scans, period) {
		// scan
		if err := singleRun(); err != nil {
			return err
		}
		// reset internal store
		if n > 0 && n%resetPeriod == 0 {
			store.Clear()
		}
	}

	return nil
}

func singleRun() error {
	// run the modules
	start := time.Now()
	if err := modules.RunModules(); err != nil {
		return err
	}
	end := time.Now()

	// collect memory performances after the scan
	perfs := perf.Collect()

	// fill the payload
	payload := store.InitPayload()
	payload.Extra = &models.ExtraInfo{
		Agent:     config.GetAgent(),
		Version:   config.Version,
		Timestamp: end,
		Duration:  end.Sub(start),
		Errors:    modules.BuildModuleErrors(),
		Perfs:     perfs,
	}

	return backends.Write(payload)
}
