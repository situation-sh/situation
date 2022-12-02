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

type fileFound struct {
	path string
}

func (m *fileFound) Error() string {
	return fmt.Sprintf("Got it! (%s)", m.path)
}

func walker(path string, d fs.DirEntry, err error) error {
	if err == nil && d.Name() == fileName {
		return &fileFound{path: path}
	}
	return nil
}

func FindDBFile() (string, error) {
	if utils.FileExists(defaultPath) {
		return defaultPath, nil
	}

	err := filepath.WalkDir(fallbackDirectory, walker)
	if ff, ok := err.(*fileFound); ok {
		return ff.path, nil
	}
	return "", fmt.Errorf("DB file not found")
}
