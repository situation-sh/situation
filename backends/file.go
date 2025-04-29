package backends

import (
	"encoding/json"
	"os"

	"github.com/situation-sh/situation/models"
	"gopkg.in/yaml.v3"
)

type FileBackend struct {
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
	RegisterBackend(b)
	// SetDefault(b, "enabled", false, "enable the backend")
	SetDefault(b, "format", &b.Format, "output format")
	SetDefault(b, "path", &b.Path, "output file")
}

func (f *FileBackend) Name() string {
	return "file"
}

func (f *FileBackend) Init() error {
	logger := GetLogger(f)
	logger.Infof("Opening file %s", f.Path)
	file, err := os.Create(f.Path)
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

	switch f.Format {
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
