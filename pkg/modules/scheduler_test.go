package modules

import (
	"testing"
)

// func TestNewScheduler(t *testing.T) {
// 	s := NewScheduler(GetEnabledModules())

// 	for n, m := range modules {
// 		if s.modules[n] != m {
// 			t.Errorf("Bad module (name: %s)", n)
// 		}
// 	}
// }

// func TestNoneMissingDependencies(t *testing.T) {
// 	s := NewScheduler(GetEnabledModules())
// 	if err := s.checkMissingDependencies(); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestMissingDependencies(t *testing.T) {
// 	m := &HostBasicModule{}
// 	if err := config.Set(disableModuleKey(m), "true"); err != nil {
// 		t.Error(err)
// 	}
// 	defer config.Set(disableModuleKey(m), "false")

// 	s := NewScheduler(GetEnabledModules())
// 	if err := s.checkMissingDependencies(); err == nil {
// 		t.Errorf("Deps must be missing")
// 	}
// }

func TestSingleRun(t *testing.T) {
	// injectDefaultConfig()
	s := NewScheduler([]Module{modules["host-basic"]})
	if err := s.Run(); err != nil {
		t.Error(err)
	}
}
