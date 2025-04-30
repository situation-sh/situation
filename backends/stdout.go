package backends

import (
	"encoding/json"
	"fmt"

	"github.com/situation-sh/situation/models"
	"gopkg.in/yaml.v3"
)

type StdoutBackend struct {
	Format string
}

func init() {
	b := &StdoutBackend{Format: jsonFormat}
	RegisterBackend(b)
	SetDefault(b, "format", &b.Format, "output format")
}

func (s *StdoutBackend) Name() string {
	return "stdout"
}

func (s *StdoutBackend) Init() error {
	return nil
}

func (s *StdoutBackend) Close() {
}

func (s *StdoutBackend) Write(p *models.Payload) {
	var bytes []byte
	var err error

	switch s.Format {
	case yamlFormat:
		bytes, err = yaml.Marshal(p)
	default:
		bytes, err = json.Marshal(p)
	}

	if err != nil {
		GetLogger(s).Error(err)
	}
	fmt.Println(string(bytes))
}
