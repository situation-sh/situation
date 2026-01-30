package modules

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// Scheduler manages the overall run of the modules
type Scheduler struct {
	modules           map[string]Module
	logger            logrus.FieldLogger
	ignoreMissingDeps bool
	supervisor        SchedulerSupervisor
	failfast          bool
}

type SchedulerSupervisor interface {
	StartChild(name string) SchedulerSupervisor
	SetStatus(err error)
	Finish()
}

type DummySchedulerSupervisor struct{}

func (s *DummySchedulerSupervisor) StartChild(name string) SchedulerSupervisor {
	return &DummySchedulerSupervisor{}
}
func (s *DummySchedulerSupervisor) Finish() {}

func (s *DummySchedulerSupervisor) SetStatus(err error) {}

type SchedulerOptions func(*Scheduler)

func WithLogger(logger logrus.FieldLogger) SchedulerOptions {
	return func(s *Scheduler) {
		s.logger = logger
	}
}

func IgnoreMissingDeps(skip bool) SchedulerOptions {
	return func(s *Scheduler) {
		s.ignoreMissingDeps = skip
	}
}

func WithSupervisor(supervisor SchedulerSupervisor) SchedulerOptions {
	return func(s *Scheduler) {
		s.supervisor = supervisor
	}
}

func FailFast() SchedulerOptions {
	return func(s *Scheduler) {
		s.failfast = true
	}
}

// NewScheduler inits a scheduler
func NewScheduler(modules []Module, options ...SchedulerOptions) *Scheduler {
	s := Scheduler{
		modules:           make(map[string]Module),
		logger:            dummyLogger(),
		ignoreMissingDeps: false,
		supervisor:        &DummySchedulerSupervisor{},
		failfast:          false,
	}
	for _, m := range modules {
		s.modules[m.Name()] = m
	}
	for _, opt := range options {
		opt(&s)
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
func (s *Scheduler) Run(ctx context.Context) error {
	// check deps
	if !s.ignoreMissingDeps {
		s.logger.Info("Checking dependencies")
		if err := s.checkMissingDependencies(); err != nil {
			return err
		}
	}

	// arrange tasks

	tasks, err := s.buildTasksList()
	if err != nil {
		return err
	}
	taskNames := make([]string, len(tasks))
	for i, t := range tasks {
		taskNames[i] = t.Name()
	}
	s.logger.WithField("tasks", taskNames).Info("Scheduling tasks")

	// set all the module status to nil
	resetStatus()

	for _, t := range tasks {
		span := s.supervisor.StartChild(t.Name())
		// run the module
		s.logger.Infof("Running module %s", t.Name())

		err := t.Run(ctx)
		if err != nil {
			if s.failfast {
				return err
			}
			s.logger.
				WithField("module", t.Name()).
				WithError(err).
				Error("Module failed")
		}
		span.SetStatus(err)
		// moduleStatus[t.Name()] = t.Run(ctx)
		span.Finish()
	}

	return nil
}
