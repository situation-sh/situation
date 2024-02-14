package cmd

import (
	"time"

	"github.com/urfave/cli/v2"

	"github.com/situation-sh/situation/backends"
	"github.com/situation-sh/situation/config"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/modules"
	"github.com/situation-sh/situation/perf"
	"github.com/situation-sh/situation/store"
)

func loopCondition(n int, scans int, period time.Duration) bool {
	if n == scans {
		return false
	}
	time.Sleep(period)
	return true
}

func runRunCmd(c *cli.Context) error {
	scans, err := config.Get[int]("scans")
	if err != nil {
		return err
	}
	period, err := config.Get[time.Duration]("period")
	if err != nil {
		return err
	}
	resetPeriod, err := config.Get[int]("reset")
	if err != nil {
		return err
	}

	if err := backends.Init(); err != nil {
		return err
	}
	defer backends.Close()

	for n, run := 0, true; run; n, run = n+1, loopCondition(n+1, scans, period) {
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

	backends.Write(payload)

	// store.Clear()
	return nil
}
