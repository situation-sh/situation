package modules

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/config"
	"github.com/situation-sh/situation/store"
)

const banner = `
███████╗██╗████████╗██╗   ██╗ █████╗ ████████╗██╗ ██████╗ ███╗   ██╗
██╔════╝██║╚══██╔══╝██║   ██║██╔══██╗╚══██╔══╝██║██╔═══██╗████╗  ██║
███████╗██║   ██║   ██║   ██║███████║   ██║   ██║██║   ██║██╔██╗ ██║
╚════██║██║   ██║   ██║   ██║██╔══██║   ██║   ██║██║   ██║██║╚██╗██║
███████║██║   ██║   ╚██████╔╝██║  ██║   ██║   ██║╚██████╔╝██║ ╚████║
╚══════╝╚═╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
███╗   ███╗ ██████╗ ██████╗ ██╗   ██╗██╗     ███████╗               
████╗ ████║██╔═══██╗██╔══██╗██║   ██║██║     ██╔════╝               
██╔████╔██║██║   ██║██║  ██║██║   ██║██║     █████╗                 
██║╚██╔╝██║██║   ██║██║  ██║██║   ██║██║     ██╔══╝                 
██║ ╚═╝ ██║╚██████╔╝██████╔╝╚██████╔╝███████╗███████╗               
╚═╝     ╚═╝ ╚═════╝ ╚═════╝  ╚═════╝ ╚══════╝╚══════╝               
████████╗███████╗███████╗████████╗██╗███╗   ██╗ ██████╗             
╚══██╔══╝██╔════╝██╔════╝╚══██╔══╝██║████╗  ██║██╔════╝             
   ██║   █████╗  ███████╗   ██║   ██║██╔██╗ ██║██║  ███╗            
   ██║   ██╔══╝  ╚════██║   ██║   ██║██║╚██╗██║██║   ██║            
   ██║   ███████╗███████║   ██║   ██║██║ ╚████║╚██████╔╝            
   ╚═╝   ╚══════╝╚══════╝   ╚═╝   ╚═╝╚═╝  ╚═══╝ ╚═════╝ 
`

func TestMain(m *testing.M) {
	// parse flags (use to test single module)
	moduleFlag := flag.String("module", "", "name of the module to run")
	flag.Parse()

	// set logrus in debug level if the verbose flag has been activated
	if f := flag.Lookup("test.v"); f != nil && f.Value.String() == "true" {
		// logrus
		logrus.SetLevel(logrus.DebugLevel)
	}

	// empty the store
	store.Clear()
	if *moduleFlag == "" {
		// run the classical test suite
		os.Exit(m.Run())
	} else {
		// run the module testing command

		// print banner
		fmt.Printf("%s", banner)

		// test a single module
		if err := testSingleModule(*moduleFlag); err != nil {
			logrus.Error(err)
			os.Exit(1)
		}
		os.Exit(0)
	}
}

// GenericTestModule is a basic function to test a module
// Developer must ensure that the store is cleared and some
// config are set
func GenericTestModule(m Module, alreadyRun map[string]bool) error {
	if alreadyRun == nil {
		alreadyRun = make(map[string]bool)
		// injectDefaultConfig()
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
			switch err.(type) {
			case *mustBeRunAsRootError:
			case *notApplicableError:
				t.Logf("warning with module %s: %v", name, err)
			default:
				t.Errorf("error with module %s: %v", name, err)
			}
		}
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
		if v {
			out = append(out, k)
		}
	}
	return out
}

// func TestGetAllDepends(t *testing.T) {
// 	mods := getAllDepends(&TCPScanModule{})
// 	for _, m := range mods {
// 		t.Log(m.Name())
// 	}
// 	s := NewScheduler(mods)
// 	if tasks, err := s.buildTasksList(); err != nil {
// 		t.Error(err)
// 	} else {
// 		for _, m := range tasks {
// 			t.Log("TASK:", m.Name())
// 		}
// 	}
// }

func testSingleModule(name string) error {
	m, ok := modules[name]
	if !ok {
		return fmt.Errorf("Module %s not found", name)
	}
	modulesToRun := getAllDepends(m)
	s := NewScheduler(modulesToRun)
	return s.Run()
}

func TestGetModuleNames(t *testing.T) {
	names := GetModuleNames()
	if len(names) != len(modules) {
		t.Errorf("bad number of modules, expect %d, got %d",
			len(modules), len(names))
	}
}

// injectDefaultConfig creates a new cli.Context
// from DefaultFlags (normally this thing is done
// by the cmd but here we are testing locally)
// func injectDefaultConfig() {
// 	fs := flag.NewFlagSet("test", flag.ExitOnError)
// 	for _, fl := range DefaultFlags {
// 		fl.Apply(fs)
// 	}
// 	c := cli.NewContext(nil, fs, nil)
// 	config.InjectContext(c)
// }

func TestGetEnabledModules(t *testing.T) {
	// injectDefaultConfig()
	em := GetEnabledModules()
	if len(em) != len(modules) {
		t.Errorf("bad number of modules, expect %d, got %d",
			len(modules), len(em))
	}
}

func TestGetEnabledModules2(t *testing.T) {
	mod := &TCPScanModule{}
	if err := config.Set(disableModuleKey(mod), "true"); err != nil {
		t.Error(err)
	}
	defer config.Set(disableModuleKey(mod), "false")

	// overrideFlag(modules[name], DISABLED_KEY, true, "")
	// defer overrideFlag(modules[name], DISABLED_KEY, false, "")

	// injectDefaultConfig()

	for _, m := range GetEnabledModules() {
		if m.Name() == mod.Name() {
			t.Errorf("Module %s should not be enabled", mod.Name())
		}
	}
}
