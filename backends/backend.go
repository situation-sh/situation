package backends

import (
	"github.com/situation-sh/situation/models"
)

var backends = make(map[string]Backend)

var enabledBackends []Backend

const jsonFormat = "json"
const yamlFormat = "yaml"

// Backend is similar to the Outputs of telegraf:
// see https://github.com/influxdata/telegraf/blob/58071a7126b4bce8862b6e75253f3ca63c9c6fc2/output.go#L3
// The backend is responsible to display errors when they occur
// It should return an error only in fatal case
type Backend interface {
	Name() string
	Init() error
	Close()
	Write(*models.Payload)
}

func isEnabled(backend Backend) bool {
	enabled, err := GetConfig[bool](backend, "enabled")
	if err == nil {
		return enabled
	}
	return false
}

// PrepareBackends select only the backend that have been enabled
func prepareBackends() {
	enabledBackends = make([]Backend, 0)
	for _, b := range backends {
		if isEnabled(b) {
			enabledBackends = append(enabledBackends, b)
		}
	}
}

// Init triggers the .Init() method of all the registered
// and enabled backends
func Init() error {
	// select only the backends that are enabled
	prepareBackends()

	for _, b := range enabledBackends {
		if err := b.Init(); err != nil {
			return err
		}
	}
	return nil
}

// Close triggers the .Close() method of all the enabled backends
func Close() {
	for _, b := range enabledBackends {
		b.Close()
	}
}

// Write triggers the .Write() method of all the enabled backends
func Write(m *models.Payload) {
	for _, b := range enabledBackends {
		b.Write(m)
	}
}
