package backends

import (
	"encoding/json"
	"os"

	"github.com/situation-sh/situation/models"
	"gopkg.in/yaml.v3"
)

type FileBackend struct {
	format string
	path   string
	file   *os.File
}

// default config
var defaultFileBackend = FileBackend{
	format: jsonFormat,
	path:   "./situation-output.json",
	file:   nil,
}

func init() {
	b := &FileBackend{format: defaultFileBackend.format}
	RegisterBackend(b)
	SetDefault(b, "enabled", false, "enable the backend")
	SetDefault(b, "format", defaultFileBackend.format, "output format")
	SetDefault(b, "path", defaultFileBackend.path, "output file")
}

func (f *FileBackend) Name() string {
	return "file"
}

func (f *FileBackend) Init() error {
	logger := GetLogger(f)

	format, err := GetConfig[string](f, "format")
	if err != nil {
		format = defaultFileBackend.format
	}
	switch format {
	case jsonFormat, yamlFormat:
		f.format = format
	default:
		f.format = jsonFormat
		// warn only (fallback to json)
		logger.Warnf(
			"Bad output format '%s'. Falling back to 'json'", format)
	}

	p, err := GetConfig[string](f, "path")
	if err != nil {
		p = defaultFileBackend.path
	}
	f.path = p

	file, err := os.Create(f.path)
	if err != nil {
		return err
	} else {
		f.file = file
	}
	return nil
}

func (f *FileBackend) Close() {
	if f.file != nil {
		if err := f.file.Close(); err != nil {
			GetLogger(f).Error(err)
		}
	}
}

func (f *FileBackend) Write(p *models.Payload) {
	logger := GetLogger(f)
	var bytes []byte
	var err error

	switch f.format {
	case yamlFormat:
		bytes, err = yaml.Marshal(p)
	default:
		bytes, err = json.Marshal(p)
	}

	if err != nil {
		logger.Error(err)
		return
	}

	_, err = f.file.Write(bytes)
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Infof("Payload written to %s", f.file.Name())
}
