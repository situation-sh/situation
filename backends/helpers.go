package backends

import (
	"fmt"

	"github.com/asiffer/puzzle"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/config"
)

func enableBackendKey(backend Backend) string {
	return fmt.Sprintf("enable-backend-%s", backend.Name())
}

func isEnabled(b Backend) bool {
	enabled, err := config.Get[bool](enableBackendKey(b))
	return err == nil && enabled
}

// GetLogger is a helper function that returns a logger specific
// to the input backend
func GetLogger(backend Backend) *logrus.Entry {
	return logrus.WithField("module", backend.Name())
}

// SetDefault is a helper that defines default backend parameter
//
// :warning: There is a bug within the lib that manages the commands and the flags
// If you define a default value as zero (false for bool, "" for string, 0 for int...)
// the value is not updated with the config file. See https://github.com/urfave/cli/issues/1395
func SetDefault[T any](backend Backend, key string, value *T, usage string) {
	name := fmt.Sprintf("backends.%s.%s", backend.Name(), key)
	flagName := fmt.Sprintf("%s-%s", backend.Name(), key)
	config.DefineVar(name, value, puzzle.WithDescription(usage), puzzle.WithFlagName(flagName))
}

// RegisterBackend is the function to call to register a backend
// It panics if two modules have the same name
func RegisterBackend(backend Backend) {
	name := backend.Name()
	if _, exists := backends[name]; exists {
		panic(fmt.Errorf("two backends have the same name: %s", name))
	}
	backends[name] = backend
	// generate a config entry
	config.Define(
		enableBackendKey(backend),
		false,
		puzzle.WithDescription(fmt.Sprintf("Enable the %s backend", backend.Name())),
		puzzle.WithFlagName(fmt.Sprintf("%s", backend.Name())),
	)
}
