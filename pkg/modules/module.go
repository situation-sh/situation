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

// type Storage struct {
// 	// store store.Store
// 	db *bun.DB
// }

// func (s Storage)

// func (s *Storage) SetStore(st store.Store) {
// 	s.store = st
// }

// type Logger struct {
// 	logger logrus.FieldLogger
// }

// func (l *Logger) SetLogger(logger logrus.FieldLogger) {
// 	l.logger = logger
// }

// func (l *Logger) GetLogger() logrus.FieldLogger {
// 	if l.logger != nil {
// 		return l.logger
// 	}
// 	// return a dummy logger
// 	return &logrus.Logger{Out: io.Discard}
// }

type BaseModule struct {
	// Storage
	// Logger
	// logger logrus.FieldLogger
	// db     *bun.DB
	// ctx    context.Context
}

// func isDisabled(m Module) bool {
// 	disabled, err := config.Get[bool](disableModuleKey(m))
// 	// if there is an error we prefer disable the module
// 	if err != nil {
// 		return true
// 	}
// 	return disabled
// }

// GetEnabledModules returns the list of the modules that
// are not disabled
// func GetEnabledModules() []Module {
// 	list := make([]Module, 0, len(modules))
// 	for _, mod := range modules {
// 		if isDisabled(mod) {
// 			continue
// 		}
// 		list = append(list, mod)
// 	}
// 	return list
// }

// RunModules does the job. This is the entrypoint
// of the "modules" sub-package. It returns an error
// only if it does not manage to schedule the modules.
// func RunModules() error {
// 	scheduler := NewScheduler(GetEnabledModules())
// 	return scheduler.Run()
// }

// func (m *BaseModule) SetLogger(logger logrus.FieldLogger) {
// 	m.logger = logger
// }

// func (m *BaseModule) SetDB(db *bun.DB) {
// 	m.db = db
// }

// func (m *BaseModule) SetValue(key any, val any) {
// 	m.ctx = context.WithValue(m.ctx, key, val)
// }

// func (m *BaseModule) GetAgent() string {
// 	v, ok := m.ctx.Value("agent").(string)
// 	if !ok {
// 		return ""
// 	}
// 	return v
// }

// func (m *BaseModule) GetHost() (*models.Machine, error) {
// 	agent := m.GetAgent()
// 	if agent == "" {
// 		return nil, fmt.Errorf("no agent in context")
// 	}
// 	machine := new(models.Machine)
// 	err := m.db.NewSelect().
// 		Model(&machine).
// 		Where("agent = ?", agent).
// 		Scan(m.ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return machine, nil
// }

// func (m *BaseModule) UpsertHost(machine *models.Machine) (*models.Machine, error) {
// 	agent := m.GetAgent()
// 	if agent == "" {
// 		return nil, fmt.Errorf("no agent in context")
// 	}
// 	machine.Agent = agent
// 	// machine := new(models.Machine)
// 	err := m.db.NewInsert().
// 		Model(&machine).
// 		On("CONFLICT (agent) DO UPDATE").
// 		Set()
// 	Scan(m.ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return machine, nil
// }
