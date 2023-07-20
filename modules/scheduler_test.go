package modules

import "testing"

func TestNewScheduler(t *testing.T) {
	injectDefaultConfig()
	s := NewScheduler(GetEnabledModules())

	for n, m := range modules {
		if s.getModuleByName(n) != m {
			t.Errorf("Bad module (name: %s)", n)
		}
	}
}

func TestMissingDependencies(t *testing.T) {
	injectDefaultConfig()
	s := NewScheduler(GetEnabledModules())
	if err := s.checkMissingDependencies(); err != nil {
		t.Error(err)
	}
}

func TestMissingDependencies2(t *testing.T) {
	overrideFlag(modules["host-network"], DISABLED_KEY, true, "")
	defer overrideFlag(modules["host-network"], DISABLED_KEY, false, "")

	injectDefaultConfig()
	s := NewScheduler(GetEnabledModules())
	if err := s.checkMissingDependencies(); err == nil {
		t.Errorf("Deps must be missing")
	}
}

func TestSingleRun(t *testing.T) {
	injectDefaultConfig()
	s := NewScheduler([]Module{modules["host-basic"]})
	if err := s.Run(); err != nil {
		t.Error(err)
	}
}
