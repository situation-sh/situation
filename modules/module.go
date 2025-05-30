package modules

import (
	"sort"

	"github.com/situation-sh/situation/config"
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

func isDisabled(m Module) bool {
	disabled, err := config.Get[bool](disableModuleKey(m))
	// if there is an error we prefer disable the module
	if err != nil {
		return true
	}
	return disabled
}

// GetEnabledModules returns the list of the modules that
// are not disabled
func GetEnabledModules() []Module {
	list := make([]Module, 0, len(modules))
	for _, mod := range modules {
		if isDisabled(mod) {
			continue
		}
		list = append(list, mod)
	}
	return list
}

// RunModules does the job. This is the entrypoint
// of the "modules" sub-package. It returns an error
// only if it does not manage to schedule the modules.
func RunModules() error {
	scheduler := NewScheduler(GetEnabledModules())
	return scheduler.Run()
}
