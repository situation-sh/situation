package backends

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/config"
	"github.com/situation-sh/situation/utils"
	"github.com/urfave/cli/v2"
)

// DefaultFlags is the list of flags that may be used to tunes the
// backends
var DefaultFlags = make([]cli.Flag, 0)

// GetLogger is a helper function that returns a logger specific
// to the input backend
func GetLogger(backend Backend) *logrus.Entry {
	return logrus.WithField("module", backend.Name())
}

// GetConfig is a generic function that returns a value
// associated to a key within the backend namespace
func GetConfig[T any](backend Backend, key string) (T, error) {
	k := fmt.Sprintf("backends.%s.%s", backend.Name(), key)
	return config.Get[T](k)
}

// SetDefault is a helper that defines default backend parameter
//
// :warning: There is a bug within the lib that manages the commands and the flags
// If you define a default value as zero (false for bool, "" for string, 0 for int...)
// the value is not updated with the config file. See https://github.com/urfave/cli/issues/1395
func SetDefault(backend Backend, key string, value interface{}, usage string) {
	name := fmt.Sprintf("backends.%s.%s", backend.Name(), key)
	if flag := utils.BuildFlag(name, value, usage, nil); flag != nil {
		DefaultFlags = append(DefaultFlags, flag)
	} else {
		panic(fmt.Errorf("cannot set default flags: %s=%v", name, value))
	}
}

// RegisterBackend is the function to call to register a backend
// It panics if two modules have the same name
func RegisterBackend(backend Backend) {
	name := backend.Name()
	if _, exists := backends[name]; exists {
		panic(fmt.Errorf("two backends have the same name: %s", name))
	}
	backends[name] = backend
}
