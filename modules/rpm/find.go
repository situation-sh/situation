//go:build linux
// +build linux

package rpm

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/situation-sh/situation/utils"
)

const fileName = "rpmdb.sqlite"
const defaultPath = "/var/lib/rpm/rpmdb.sqlite"
const fallbackDirectory = "/usr/lib"

type FileFound struct {
	path string
}

func (m *FileFound) Error() string {
	return fmt.Sprintf("Got it! (%s)", m.path)
}

func walker(path string, d fs.DirEntry, err error) error {
	if err == nil && d.Name() == fileName {
		return &FileFound{path: path}
	}
	return nil
}

func FindDBFile() (string, error) {
	if utils.FileExists(defaultPath) {
		return defaultPath, nil
	}

	err := filepath.WalkDir(fallbackDirectory, walker)
	if ff, ok := err.(*FileFound); ok {
		return ff.path, nil
	}
	return "", fmt.Errorf("DB file not found")
}
