package modules

import "github.com/situation-sh/situation/pkg/models"

// moduleStatus stores the status of the run modules
var moduleStatus map[string]error

func resetStatus() {
	moduleStatus = make(map[string]error)
	for _, m := range mods {
		moduleStatus[m.Name()] = nil
	}
}

// BuildModuleErrors wraps all errors sent by module into
// a single list of ModuleError
func BuildModuleErrors() []*models.ModuleError {
	list := make([]*models.ModuleError, 0)
	for mod, err := range moduleStatus {
		if err != nil {
			list = append(list, &models.ModuleError{Module: mod, Message: err.Error()})
		}
	}
	return list
}
