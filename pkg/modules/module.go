package modules

import (
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/pkg/store"
)

// internal map of modules
var modules = make(map[string]Module)

// Module is the generic module interface to implement plugins to
// the agent
type Module interface {
	Name() string
	Dependencies() []string
	Run() error
}

// GetModuleNames return the list of all the available modules
func GetModuleNames() []string {
	list := make([]string, len(modules))
	i := 0
	for name := range modules {
		list[i] = name
		i++
	}
	// sort the module names
	sort.Strings(list)
	return list
}

func GetModuleByName(name string) Module {
	return modules[name]
}

func Walk(fun func(name string, mod Module)) {
	for name, mod := range modules {
		fun(name, mod)
	}
}

type Storage struct {
	store store.Store
}

func (s *Storage) SetStore(st store.Store) {
	s.store = st
}

type Logger struct {
	logger logrus.FieldLogger
}

func (l *Logger) SetLogger(logger logrus.FieldLogger) {
	l.logger = logger
}

// func (l *Logger) GetLogger() logrus.FieldLogger {
// 	if l.logger != nil {
// 		return l.logger
// 	}
// 	// return a dummy logger
// 	return &logrus.Logger{Out: io.Discard}
// }

type BaseModule struct {
	Storage
	Logger
}

// func isDisabled(m Module) bool {
// 	disabled, err := config.Get[bool](disableModuleKey(m))
// 	// if there is an error we prefer disable the module
// 	if err != nil {
// 		return true
// 	}
// 	return disabled
// }

// GetEnabledModules returns the list of the modules that
// are not disabled
// func GetEnabledModules() []Module {
// 	list := make([]Module, 0, len(modules))
// 	for _, mod := range modules {
// 		if isDisabled(mod) {
// 			continue
// 		}
// 		list = append(list, mod)
// 	}
// 	return list
// }

// RunModules does the job. This is the entrypoint
// of the "modules" sub-package. It returns an error
// only if it does not manage to schedule the modules.
// func RunModules() error {
// 	scheduler := NewScheduler(GetEnabledModules())
// 	return scheduler.Run()
// }
