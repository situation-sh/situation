package modules

import (
	"fmt"

	"github.com/asiffer/puzzle"
)

func disableModuleKey(m Module) string {
	return fmt.Sprintf("disable-module-%s", m.Name())
}

// GetLogger is a helper function that returns a logger specific
// to the input module
//
//	func GetLogger(m Module) *logrus.Entry {
//		return logrus.WithField("module", m.Name())
//	}
// func GetLogger(m BaseModule) logrus.FieldLogger {
// 	if m.logger != nil {
// 		return m.logger
// 	}
// 	return &logrus.Entry{}
// }

// SetDefault is a helper that defines default module parameter.
// The provided values can be overwritten by CLI flags, env variables or anything
// the asiffer/puzzle library may support.
func setDefault[T any](config *puzzle.Config, m Module, key string, value *T, usage string) error {
	name := fmt.Sprintf("modules.%s.%s", m.Name(), key)
	return puzzle.DefineVar(
		config,
		name,
		value,
		puzzle.WithDescription(usage),
		puzzle.WithFlagName(fmt.Sprintf("%s-%s", m.Name(), key)))
}

// registerModule is the function to call to register a module
// It panics if two modules have the same name
func registerModule(module Module) {
	name := module.Name()
	if _, exists := modules[name]; exists {
		panic(fmt.Errorf("two modules have the same name: %s", name))
	}
	modules[name] = module
	// config.Define(
	// 	disableModuleKey(module),
	// 	false,
	// 	puzzle.WithDescription(fmt.Sprintf("Disable module %s", name)),
	// 	puzzle.WithFlagName(fmt.Sprintf("no-module-%s", name)),
	// )
}
