package modules

import (
	"flag"
	"testing"

	"github.com/situation-sh/situation/config"
)

var singleModule string


func init() {
	flag.StringVar(&singleModule, "module", "", "name of the module to run")
	if err := config.PopulateFlags(flag.CommandLine); err != nil {
		panic(err)
	}
}

func TestAllModules(t *testing.T) {
	if singleModule != "" {
		// run a single module in priority
		mod, ok := modules[singleModule]
		if !ok {
			t.Fatalf("Module %s not found", singleModule)
		}
		if err := testSingleModule(mod); err != nil {
			t.Error(err)
		}
		return
	}

	// run all the enabled modules
	for _, mod := range GetEnabledModules() {
		t.Run(mod.Name(), func(t *testing.T) {
			if err := testSingleModule(mod); err != nil {
				t.Error(err)
			}
		})
	}
}

// getAllDepends return all the modules required to run a given module
func getAllDepends(m Module) []Module {
	results := map[Module]bool{m: true}
	for _, name := range m.Dependencies() {
		if d, ok := modules[name]; ok {
			for _, n := range getAllDepends(d) {
				results[n] = true
			}
		}
	}
	out := make([]Module, 0)
	for k, v := range results {
		// prune disabled modules in the end
		if v && !isDisabled(k) {
			out = append(out, k)
		}
	}
	return out
}

func testSingleModule(m Module) error {
	modulesToRun := getAllDepends(m)
	s := NewScheduler(modulesToRun)
	return s.Run()
}

