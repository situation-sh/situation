package cmd

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v3"
)

var (
	startTime    time.Time     = time.Now()
	daysPeriod   uint          = 1
	timePeriod   time.Duration = 0
	uninstall    bool          = false
	taskLogLevel uint          = 1
)

var taskCmd = cli.Command{
	Name:    "task",
	Aliases: []string{"cron"},
	Usage:   "Install a scheduled task",
	Action:  runTaskCmd,
	// SkipFlagParsing: true, // to allow extra flags to be passed to the run command
	Flags: []cli.Flag{
		&cli.TimestampFlag{
			Name:        "task-start",
			Usage:       "Start time of the task (default: now). Use either '15:04:05' or '3:04PM' format",
			Value:       time.Now(),
			Destination: &startTime,
			Config:      cli.TimestampConfig{Layouts: []string{time.TimeOnly, time.Kitchen}},
			HideDefault: true,
		},
		&cli.UintFlag{
			Name:        "task-days",
			Usage:       "Interval between two runs in days (default: 1, i.e. daily)",
			Value:       1,
			Destination: &daysPeriod,
		},
		&cli.DurationFlag{
			Name:        "task-period",
			Usage:       "Interval between two runs within one day (e.g. '30m' for every 30 minutes, disabled by default)",
			Value:       0,
			Destination: &timePeriod,
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "uninstall",
			Usage:       "Uninstall the scheduled task",
			Value:       false,
			Destination: &uninstall,
		},
		&cli.UintFlag{
			Name:        "task-log-level",
			Usage:       "Define the log level used in the scheduled task (0: Panic, 1: Fatal, 2: Error, 3: Warn, 4: Info, 5: Debug)",
			Value:       1,
			Destination: &taskLogLevel,
		},
	},
}

func init() {
	// inject run flags
	taskCmd.Flags = append(taskCmd.Flags, runCmd.Flags...)
}

func getRunArgs(cmd *cli.Command) []string {
	out := make([]string, 0)
	for _, name := range runCmd.FlagNames() { // only root-defined (global) flags
		// fmt.Println("name:", name)
		if len(name) == 1 {
			// skip short names
			// corresponding long names are also in the list
			// it seems to be the behaviour of flag.FlagSet
			continue
		}
		if name == "log-level" {
			continue // we set it specifically for the task with task-log-level
		}
		if cmd.IsSet(name) { // user provided it
			// Value() returns a cli.Value; String() gives a normalized string form
			switch v := cmd.Value(name).(type) {
			case bool:
				out = append(out, fmt.Sprintf("--%s", name))
			default:
				out = append(out, fmt.Sprintf("--%s", name), fmt.Sprintf("%v", v))
			}

		}
	}
	return out
}
