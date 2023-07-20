package modules

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/config"
	"github.com/situation-sh/situation/utils"
	"github.com/urfave/cli/v2"
)

// DefaultFlags is the list of flags that may be used to tunes the
// modules
var DefaultFlags = make([]cli.Flag, 0)

// GetLogger is a helper function that returns a logger specific
// to the input module
func GetLogger(m Module) *logrus.Entry {
	return logrus.WithField("module", m.Name())
}

// GetConfig is a generic function that returns a value
// associated to a key within the module namespace
func GetConfig[T any](m Module, key string) (T, error) {
	k := fmt.Sprintf("modules.%s.%s", m.Name(), key)
	return config.Get[T](k)
}

// SetDefault is a helper that defines default module parameter.
// The provided values can be overwritten by CLI flags or config file.
func SetDefault(m Module, key string, value interface{}, usage string) {
	name := fmt.Sprintf("modules.%s.%s", m.Name(), key)
	if flag := utils.BuildFlag(name, value, usage, nil); flag != nil {
		DefaultFlags = append(DefaultFlags, flag)
	}
}

// RegisterModule is the function to call to register a module
// It panics if two modules have the same name
func RegisterModule(module Module) {
	name := module.Name()
	if _, exists := modules[name]; exists {
		panic(fmt.Errorf("two modules have the same name: %s", name))
	}
	modules[name] = module
	// add a default parameter to disable the module
	SetDefault(module, DISABLED_KEY, false, fmt.Sprintf("Disable module %s", name))
}
