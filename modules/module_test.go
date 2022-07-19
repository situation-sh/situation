package modules

import (
	"fmt"
	"os"
	"testing"

	"github.com/situation-sh/situation/store"
)

func TestMain(m *testing.M) {
	// empty the store
	store.Clear()
	os.Exit(m.Run())
}

// GenericTestModule is a basic function to test a module
// Developer must ensure that the store is cleared and some
// config are set
func GenericTestModule(m Module, alreadyRun map[string]bool) error {
	if alreadyRun == nil {
		alreadyRun = make(map[string]bool)
	}
	// run dependencies
	for _, name := range m.Dependencies() {
		dep := modules[name]
		depName := dep.Name()

		if !alreadyRun[depName] {
			if err := GenericTestModule(dep, alreadyRun); err != nil {
				return err
			}
			alreadyRun[depName] = true
		}

	}
	// now run the module
	return m.Run()
}

func TestModules(t *testing.T) {
	for name, m := range modules {
		// empty the store
		store.Clear()
		// config.InitConfig()
		fmt.Printf("--- MODULE: %s\n", name)
		if err := GenericTestModule(m, nil); err != nil {
			t.Errorf("error with module %s: %v", name, err)
		}
	}
}
