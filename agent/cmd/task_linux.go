//go:build linux

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

const cronFile = "/etc/cron.d/situation"

func runTaskCmd(ctx context.Context, cmd *cli.Command) error {
	if uninstall {
		logrus.Infof("Removing cron job file %s", cronFile)
		return os.Remove(cronFile)
	}
	file, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot get executable path: %w", err)
	}
	// Example of job definition:
	// .---------------- minute (0 - 59)
	// |  .------------- hour (0 - 23)
	// |  |  .---------- day of month (1 - 31)
	// |  |  |  .------- month (1 - 12) OR jan,feb,mar,apr ...
	// |  |  |  |  .---- day of week (0 - 6) (Sunday=0 or 7) OR sun,mon,tue,wed,thu,fri,sat
	// |  |  |  |  |
	// *  *  *  *  *  user command to be executed
	minutes := fmt.Sprintf("%d", startTime.Minute())
	hours := fmt.Sprintf("%d", startTime.Hour())
	day := "*"
	if daysPeriod > 1 {
		day = fmt.Sprintf("*/%d", daysPeriod)
	}
	if timePeriod.Hours() >= 1.0 {
		hours = fmt.Sprintf("*/%0.f", timePeriod.Hours())
	} else if timePeriod.Minutes() > 0.0 {
		minutes = fmt.Sprintf("*/%0.f", timePeriod.Minutes())
	}
	cronLine := fmt.Sprintf("%s %s %s * * %s %s\n", minutes, hours, day, file, strings.Join(getRunArgs(cmd), " "))

	logrus.Infof("Creating cron job file %s", cronFile)
	f, err := os.OpenFile(cronFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("cannot open or create %s: %w", cronFile, err)
	}
	defer f.Close()

	logrus.Debugf("Writing cron line: %s", cronLine)
	if _, err := f.WriteString(cronLine); err != nil {
		return fmt.Errorf("cannot write to %s: %w", cronFile, err)
	}
	return nil
}
