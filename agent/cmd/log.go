package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

var begin time.Time

type ModuleFormatter struct{}

var faint = color.New(color.Faint).SprintFunc()

// var info = color.New(color.FgGreen, color.BgBlue).SprintFunc()

// var info = color.New(color.FgRed, color.BgWhite).Sprintf("%s", "INF")
var debug = color.New(color.Reset).Sprintf("%s", "DBG")
var info = color.New(color.FgBlue).Sprintf("%s", "INF")
var warning = color.New(color.FgYellow).Sprintf("%s", "WRN")
var err = color.New(color.BgRed).Sprintf("%s", "ERR")

// func Faintf(format string, a ...interface{}) string {
// 	return color.Faint(fmt.Sprintf(format, a...))
// }

func getLevelString(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel:
		return debug
	case logrus.InfoLevel:
		return info
	case logrus.WarnLevel:
		return warning
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return err
	default:
		return info
	}
}

func (f *ModuleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	d := faint(fmt.Sprintf("%010.6f", entry.Time.Sub(begin).Seconds()))
	// msg := info(entry.Message)
	level := getLevelString(entry.Level)
	module := ""
	if m, ok := entry.Data["module"]; ok {
		module = fmt.Sprintf("%14s", m)
	} else {
		module = fmt.Sprintf("%14s", "")
	}
	base := []string{d, level, faint(module), entry.Message}
	for k, v := range entry.Data {
		if k != "module" {
			q := strconv.Quote(fmt.Sprintf("%v", v))
			base = append(base, fmt.Sprintf("%s%s", faint(k+"="), q))
		}
	}
	s := strings.Join(append(base, "\n"), " ")
	return []byte(s), nil
}
