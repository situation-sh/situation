package modules

import (
	"context"
	"sort"
)

// internal map of modules
var mods = make(map[string]Module)

// Module is the generic module interface to implement plugins to
// the agent
type Module interface {
	Name() string
	Dependencies() []string
	Run(ctx context.Context) error
}

// GetModuleNames return the list of all the available modules
func GetModuleNames() []string {
	list := make([]string, len(mods))
	i := 0
	for name := range mods {
		list[i] = name
		i++
	}
	// sort the module names
	sort.Strings(list)
	return list
}

func GetModuleByName(name string) Module {
	return mods[name]
}

func Walk(fun func(name string, mod Module)) {
	for name, mod := range mods {
		fun(name, mod)
	}
}

// BaseModule is a struct that can be embedded in other modules to provide
// common functionality. It doesn't implement the Module interface itself, so
type BaseModule struct{}
