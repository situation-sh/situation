package backends

import (
	"encoding/json"
	"fmt"

	"github.com/situation-sh/situation/models"
	"gopkg.in/yaml.v3"
)

type StdoutBackend struct {
	format string
}

// default config
var defaultStdoutBackend = StdoutBackend{format: jsonFormat}

func init() {
	b := &StdoutBackend{format: defaultStdoutBackend.format}
	RegisterBackend(b)
	SetDefault(b, "enabled", true, "enable the backend")
	SetDefault(b, "format", defaultStdoutBackend.format, "output format")
}

func (s *StdoutBackend) Name() string {
	return "stdout"
}

func (s *StdoutBackend) Init() error {
	logger := GetLogger(s)

	format, err := GetConfig[string](s, "format")
	if err != nil {
		format = defaultStdoutBackend.format
	}

	switch format {
	case jsonFormat, yamlFormat:
		s.format = format
	default:
		s.format = jsonFormat
		// warn only (fallback to json)
		logger.Warnf(
			"Bad output format '%s'. Falling back to 'json'", format)
	}
	return nil
}

func (s *StdoutBackend) Close() {
}

func (s *StdoutBackend) Write(p *models.Payload) {
	var bytes []byte
	var err error

	switch s.format {
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
