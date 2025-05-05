package modules

import (
	"fmt"

	"github.com/asiffer/puzzle"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/config"
)

var skipMissingDeps = false

func init() {
	config.DefineVar(
		"skip-missing-deps",
		&skipMissingDeps,
		puzzle.WithDescription("Skip missing dependencies"),
	)
}

// Scheduler manages the overall run of the modules
type Scheduler struct {
	modules map[string]Module
	logger  *logrus.Entry
}

// NewScheduler inits a scheduler
func NewScheduler(modules []Module) *Scheduler {
	s := Scheduler{
		modules: make(map[string]Module),
		logger:  logrus.WithField("from", "scheduler"),
	}
	for _, m := range modules {
		s.modules[m.Name()] = m
	}
	return &s
}

func (s *Scheduler) checkMissingDependencies() error {
	for _, mod := range s.modules {
		for _, d := range mod.Dependencies() {
			if _, exists := s.modules[d]; exists == false {
				return fmt.Errorf("module %s needs %s which is missing", mod.Name(), d)
			}
		}
	}
	return nil
}

func (s *Scheduler) actualDependencies(m Module) []string {
	// get the dependencies of the module
	out := make([]string, 0)
	// ignore the modules that are not scheduled
	for _, dep := range m.Dependencies() {
		if _, exists := s.modules[dep]; exists {
			out = append(out, dep)
		}
	}
	return out
}

func (s *Scheduler) children(name string) []string {
	out := make([]string, 0)
	for _, m := range s.modules {
		for _, dep := range s.actualDependencies(m) {
			if dep == name {
				out = append(out, m.Name())
			}
		}
	}
	return out
}

// buildTasksList builds the list of tasks to run
// in the right order. It uses a DFS algorithm to
// traverse the graph of dependencies. It returns
// an error if there is a cycle in the graph.
func (s *Scheduler) buildTasksList() ([]Module, error) {
	var visit func(string) error
	tasks := make([]Module, 0)

	notPermanents := make([]string, 0)
	temporary := make(map[string]bool)

	for _, m := range s.modules {
		notPermanents = append(notPermanents, m.Name())
	}

	isPermanent := func(m string) bool {
		// check if the module is permanent
		for _, np := range notPermanents {
			if np == m {
				return false
			}
		}
		return true
	}

	markAsPermanent := func(m string) {
		// mark the module as permanent
		for i, np := range notPermanents {
			if np == m {
				// remove it from the list
				notPermanents = append(notPermanents[:i], notPermanents[i+1:]...)
				break
			}
		}
	}

	visit = func(m string) error {
		if isPermanent(m) {
			return nil
		}
		if _, ok := temporary[m]; ok {
			// cycle detected
			return fmt.Errorf("module dependency error (there is probably a cycle)")
		}
		temporary[m] = true

		for _, child := range s.children(m) {
			if err := visit(child); err != nil {
				return err
			}
		}

		markAsPermanent(m)
		// preprend the module to the list
		tasks = append([]Module{s.modules[m]}, tasks...)
		return nil
	}

	for len(notPermanents) > 0 {
		if err := visit(notPermanents[0]); err != nil {
			return tasks, err
		}
	}

	return tasks, nil
}

// Run returns an error only if the scheduler fails to
// plan the modules. It does not return error if a module fails
func (s *Scheduler) Run() error {
	// check deps
	if !skipMissingDeps {
		s.logger.Info("Checking dependencies")
		if err := s.checkMissingDependencies(); err != nil {
			return err
		}
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
