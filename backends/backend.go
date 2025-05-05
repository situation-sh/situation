package backends

import (
	"errors"
	"fmt"

	"github.com/situation-sh/situation/models"
)

var backends = make(map[string]Backend)

const jsonFormat = "json"
const yamlFormat = "yaml"

// Backend is similar to the Outputs of telegraf:
// see https://github.com/influxdata/telegraf/blob/58071a7126b4bce8862b6e75253f3ca63c9c6fc2/output.go#L3
// The backend is responsible to display errors when they occur
// It should return an error only in fatal case
type Backend interface {
	Name() string
	Init() error
	Close() error
	Write(*models.Payload) error
}

// Init triggers the .Init() method of all the registered
// and enabled backends
func Init() error {
	// select only the backends that are enabled
	for _, b := range backends {
		if !isEnabled(b) {
			continue
		}
		if err := b.Init(); err != nil {
			return err
		}
	}
	return nil
}

// Close triggers the .Close() method of all the enabled backends
func Close() error {
	var errs []error

	for _, b := range backends {
		if !isEnabled(b) {
			continue
		}
		if err := b.Close(); err != nil {
			wrapped := fmt.Errorf("backend %s failed to close: %w", b.Name(), err)
			errs = append(errs, wrapped)
		}
	}

	return errors.Join(errs...)
}

// Write triggers the .Write() method of all the enabled backends
func Write(m *models.Payload) error {
	var errs []error

	for _, b := range backends {
		if !isEnabled(b) {
			continue
		}
		if err := b.Write(m); err != nil {
			wrapped := fmt.Errorf("backend %s failed to write: %w", b.Name(), err)
			errs = append(errs, wrapped)
		}
	}

	return errors.Join(errs...)
}
