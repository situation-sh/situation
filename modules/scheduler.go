package modules

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// Scheduler manages the overall run of the modules
type Scheduler struct {
	modules []Module
	logger  *logrus.Entry
}

// NewScheduler inits a scheduler
func NewScheduler(modules []Module) *Scheduler {
	return &Scheduler{
		modules: modules,
		logger:  logrus.WithField("from", "scheduler"),
	}
}

func (s *Scheduler) getModuleByName(name string) Module {
	for _, mod := range modules {
		if mod.Name() == name {
			return mod
		}
	}
	return nil
}

func (s *Scheduler) checkMissingDependencies() error {
	for _, mod := range s.modules {
		for _, d := range mod.Dependencies() {
			if s.getModuleByName(d) == nil {
				return fmt.Errorf("module %s needs %s which is missing", mod.Name(), d)
			}
		}
	}
	return nil
}

func (s *Scheduler) buildTasksList() ([]Module, error) {
	tasks := make([]Module, 0)
	alreadyAdded := make(map[string]bool)
	n := len(s.modules)

	canBeAddedToTasksList := func(m Module) bool {
		for _, d := range m.Dependencies() {
			if _, exists := alreadyAdded[d]; !exists {
				return false
			}
		}
		return true
	}

	for len(tasks) < n {
		i := 0
		for _, m := range modules {
			if _, ok := alreadyAdded[m.Name()]; ok {
				continue
			}
			// add it if possible
			if canBeAddedToTasksList(m) {
				alreadyAdded[m.Name()] = true
				tasks = append(tasks, m)
				i++
			}
		}
		// it means that among all the modules no one can be added
		if i == 0 {
			return tasks, fmt.Errorf("module dependency error (there is probably a cycle)")
		}
	}
	return tasks, nil
}

// Run returns an error only if the scheduler fails to
// plan the modules. It does not return error if a module fails
func (s *Scheduler) Run() error {
	// check deps
	s.logger.Info("Checking dependencies")
	if err := s.checkMissingDependencies(); err != nil {
		return err
	}

	// arrange tasks
	s.logger.Info("Scheduling tasks")
	tasks, err := s.buildTasksList()
	if err != nil {
		return err
	}

	// set all the module status to nil
	resetStatus()

	for _, t := range tasks {
		// run the module
		s.logger.Infof("Running module %s", t.Name())
		moduleStatus[t.Name()] = t.Run()
	}

	return nil
}
