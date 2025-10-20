package backends

import (
	"encoding/json"
	"fmt"

	"github.com/asiffer/puzzle"
	"github.com/situation-sh/situation/pkg/models"
	"gopkg.in/yaml.v3"
)

type StdoutBackend struct {
	BaseBackend

	Format string
}

func init() {
	b := &StdoutBackend{Format: jsonFormat}
	registerBackend(b)
}

func (s *StdoutBackend) Name() string {
	return "stdout"
}

func (s *StdoutBackend) Bind(config *puzzle.Config) error {
	if err := setDefault(config, s, "format", &s.Format, "output format"); err != nil {
		return err
	}
	return nil
}

func (s *StdoutBackend) Init() error {
	return nil
}

func (s *StdoutBackend) Close() error {
	return nil
}

func (s *StdoutBackend) Write(p *models.Payload) error {
	var bytes []byte
	var err error

	switch s.Format {
	case yamlFormat:
		bytes, err = yaml.Marshal(p)
	default:
		bytes, err = json.Marshal(p)
	}

	if err != nil {
		return fmt.Errorf("error while marshalling payload: %w", err)
	}
	fmt.Println(string(bytes))
	return nil
}
