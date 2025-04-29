package modules

import (
	"testing"

	"github.com/situation-sh/situation/config"
)

func TestNewScheduler(t *testing.T) {
	// injectDefaultConfig()
	s := NewScheduler(GetEnabledModules())

	for n, m := range modules {
		if s.getModuleByName(n) != m {
			t.Errorf("Bad module (name: %s)", n)
		}
	}
}

func TestMissingDependencies(t *testing.T) {
	// injectDefaultConfig()
	s := NewScheduler(GetEnabledModules())
	if err := s.checkMissingDependencies(); err != nil {
		t.Error(err)
	}
}

func TestMissingDependencies2(t *testing.T) {
	m := &HostBasicModule{}
	if err := config.Set(disableModuleKey(m), "true"); err != nil {
		t.Error(err)
	}
	defer config.Set(disableModuleKey(m), "false")

	s := NewScheduler(GetEnabledModules())
	if err := s.checkMissingDependencies(); err == nil {
		t.Errorf("Deps must be missing")
	}
}

func TestSingleRun(t *testing.T) {
	// injectDefaultConfig()
	s := NewScheduler([]Module{modules["host-basic"]})
	if err := s.Run(); err != nil {
		t.Error(err)
	}
}
