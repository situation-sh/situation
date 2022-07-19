package utils

import (
	"time"

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

var internalFlags = make(map[string]interface{})

// BuildFlag returns a Flag that can be used for urfave/cli app
func BuildFlag(key string, value interface{}, help string, aliases []string) cli.Flag {
	internalFlags[key] = value

	switch v := value.(type) {
	case bool:
		return altsrc.NewBoolFlag(&cli.BoolFlag{Name: key, Value: v, Usage: help, Aliases: aliases})
	case int:
		return altsrc.NewIntFlag(&cli.IntFlag{Name: key, Value: v, Usage: help, Aliases: aliases})
	case int64:
		return altsrc.NewInt64Flag(&cli.Int64Flag{Name: key, Value: v, Usage: help, Aliases: aliases})
	case string:
		return altsrc.NewStringFlag(&cli.StringFlag{Name: key, Value: v, Usage: help, Aliases: aliases})
	case time.Duration:
		return altsrc.NewDurationFlag(&cli.DurationFlag{Name: key, Value: v, Usage: help, Aliases: aliases})
	case []string:
		return altsrc.NewStringSliceFlag(
			&cli.StringSliceFlag{Name: key, Value: cli.NewStringSlice(v...), Usage: help, Aliases: aliases})
	default:
		delete(internalFlags, key)
		// do not append
		return nil
	}
}

func BuiltFlags() map[string]interface{} {
	return internalFlags
}
