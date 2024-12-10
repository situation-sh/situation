package config

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
)

var context *cli.Context

// InjectContext receives the urfave/cli
// context to manage configuration
func InjectContext(c *cli.Context) {
	context = c
}

func Get[T any](key string) (T, error) {
	var test T
	ok := false
	var value interface{}

	switch any(test).(type) {
	case []string:
		value = interface{}(context.StringSlice(key))
	case []int:
		value = interface{}(context.IntSlice(key))
	case []int64:
		value = interface{}(context.Int64Slice(key))
	case time.Duration:
		value = interface{}(context.Duration(key))
	default:
		value = context.Value(key)
	}

	// fmt.Printf("[%s] %v (%T)\n", key, value, value)
	typedValue, ok := (value).(T)
	if ok {
		return typedValue, nil
	}
	return typedValue, fmt.Errorf("type casting has failed for key '%s' (%T)", key, value)
}
