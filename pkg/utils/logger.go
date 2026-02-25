package utils

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/agent/config"
)

type ModuleFormatter struct {
	start time.Time
}

var faint = color.New(color.Faint).SprintFunc()

var debug = color.New(color.Reset).Sprintf("%s", "DBG")
var info = color.New(color.FgBlue).Sprintf("%s", "INF")
var warning = color.New(color.FgYellow).Sprintf("%s", "WRN")
var err = color.New(color.BgRed).Sprintf("%s", "ERR")

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

type kv struct {
	key   string
	value any
}

func extractData(entry *logrus.Entry) (string, string, []kv) {
	others := make([]kv, 0)
	module := "system"
	if m, ok := entry.Data["module"]; ok {
		module = fmt.Sprintf("%s", m)
	}
	for k, v := range entry.Data {
		if k != "module" {
			others = append(others, kv{key: k, value: v})
		}
	}
	sort.Slice(others, func(i, j int) bool {
		return others[i].key < others[j].key
	})
	loc := ""
	if entry.Caller != nil {
		loc = fmt.Sprintf("%s:%d", strings.TrimPrefix(entry.Caller.File, config.Module+"/"), entry.Caller.Line)
		// others = append(others, kv{key: "loc", value: loc})
		// others = append(others, kv{key: "file", value: strings.TrimPrefix(entry.Caller.File, config.Module+"/")})
		// others = append(others, kv{key: "line", value: entry.Caller.Line})
	}
	return module, loc, others
}

func (f *ModuleFormatter) Format(entry *logrus.Entry) ([]byte, error) {

	d := faint(fmt.Sprintf("%010.6f", entry.Time.Sub(f.start).Seconds()))

	level := getLevelString(entry.Level)
	module, loc, data := extractData(entry)

	base := []string{d, level}
	if loc != "" {
		base = append(base, fmt.Sprintf("%-31s", loc))
	}
	base = append(base, faint(fmt.Sprintf("%13s", module)), fmt.Sprintf("%-36s", entry.Message))
	for _, kv := range data {
		q := strconv.Quote(fmt.Sprintf("%v", kv.value))
		base = append(base, fmt.Sprintf("%s%s", faint(kv.key+"="), q))
	}
	s := strings.Join(append(base, "\n"), " ")
	return []byte(s), nil
}

func inferOutput() io.Writer {
	if runtime.GOOS == "windows" {
		// Colored output of logrus does not work for windows
		// but we can circumvent it with ansi color codes
		// https://github.com/sirupsen/logrus/issues/172
		return ansicolor.NewAnsiColorWriter(os.Stderr)
	} else {
		return os.Stderr
		// logrus.SetOutput(output)
	}
}

func NewLogger() *logrus.Logger {
	logger := logrus.Logger{
		Out:       inferOutput(),
		Formatter: &ModuleFormatter{start: time.Now()},
		Level:     logrus.InfoLevel,
	}
	return &logger
}
