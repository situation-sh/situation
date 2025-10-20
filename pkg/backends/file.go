package backends

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/asiffer/puzzle"
	"github.com/situation-sh/situation/pkg/models"
	"gopkg.in/yaml.v3"
)

type FileBackend struct {
	BaseBackend

	Format string
	Path   string
	file   *os.File
}

func init() {
	b := &FileBackend{
		Format: jsonFormat,
		Path:   "situation.json",
		file:   nil,
	}
	registerBackend(b)
	// SetDefault(b, "enabled", false, "enable the backend")
	// SetDefault(b, "format", &b.Format, "output format")
	// SetDefault(b, "path", &b.Path, "output file")
}

func (f *FileBackend) Name() string {
	return "file"
}

func (f *FileBackend) Bind(config *puzzle.Config) error {
	if err := setDefault(config, f, "format", &f.Format, "output format"); err != nil {
		return err
	}
	if err := setDefault(config, f, "path", &f.Path, "output file"); err != nil {
		return err
	}
	return nil
}

func (f *FileBackend) Init() error {
	f.logger.Infof("Opening file %s", f.Path)
	file, err := os.Create(f.Path)
	if err != nil {
		return err
	} else {
		f.file = file
	}
	return nil
}

func (f *FileBackend) Close() error {
	if f.file != nil {
		if err := f.file.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileBackend) Write(p *models.Payload) error {
	var bytes []byte
	var err error

	switch f.Format {
	case yamlFormat:
		bytes, err = yaml.Marshal(p)
	default:
		bytes, err = json.Marshal(p)
	}

	if err != nil {
		return fmt.Errorf("error while marshalling payload: %w", err)
	}

	_, err = f.file.Write(bytes)
	if err != nil {
		return err
	}
	f.logger.Infof("Payload written to %s", f.file.Name())
	return nil
}
