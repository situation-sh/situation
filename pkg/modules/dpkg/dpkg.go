package dpkg

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/utils"
)

const (
	logDir        = "/var/log"
	logFilePrefix = "dpkg.log"
	dpkgDir       = "/var/lib/dpkg/info"
	logFormat     = "2006-01-02 15:04:05" // see https://pkg.go.dev/time#pkg-constants
	manager       = "dpkg"
)

// GetFiles returns the files installed by a package given its name
func GetFiles(name string) ([]string, error) {
	entries, err := os.ReadDir(dpkgDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), name) && strings.HasSuffix(entry.Name(), ".list") {
			lines, err := utils.GetLines(filepath.Join(dpkgDir, entry.Name()), filepath.Clean)
			if err != nil {
				return nil, err
			}
			return utils.KeepLeaves(lines), nil
		}
	}
	return nil, fmt.Errorf("no file found for %s", name)
}

func parseLogLine(line string) *models.Package {
	// Examples of lines
	// 2022-11-29 08:46:43 status unpacked linux-generic:amd64 5.15.0.53.53
	// 2022-11-29 08:46:43 status half-configured linux-generic:amd64 5.15.0.53.53
	// 2022-11-29 08:46:43 status installed linux-generic:amd64 5.15.0.53.53
	// 2022-11-29 08:46:43 trigproc libc-bin:amd64 2.35-0ubuntu3.1 <none>
	// 2022-11-29 08:46:43 status half-configured libc-bin:amd64 2.35-0ubuntu3.1
	// 2022-11-29 08:46:43 status installed libc-bin:amd64 2.35-0ubuntu3.1
	// 2022-11-29 08:46:43 trigproc initramfs-tools:all 0.140ubuntu13 <none>
	// 2022-11-29 08:46:43 status half-configured initramfs-tools:all 0.140ubuntu13
	// 2022-11-29 08:47:02 status installed initramfs-tools:all 0.140ubuntu13
	// 2022-11-29 08:47:02 trigproc linux-image-5.15.0-53-generic:amd64 5.15.0-53.59 <none>
	// 2022-11-29 08:47:02 status half-configured linux-image-5.15.0-53-generic:amd64 5.15.0-53.59

	chunks := strings.Split(line, " ")
	// filter lines that has not the right format
	if len(chunks) < 6 {
		return nil
	}
	// keep only the installed logs
	if !(chunks[2] == "status" && chunks[3] == "installed") {
		return nil
	}
	// try to parse time
	t, err := time.Parse(logFormat, fmt.Sprintf("%s %s", chunks[0], chunks[1]))
	if err != nil {
		return nil
	}
	// get name and version
	name := strings.Split(chunks[4], ":")[0] // remove the arch
	version := strings.Split(chunks[5], "-")[0]
	return &models.Package{
		Name:            name,
		Version:         version,
		Manager:         manager,
		InstallTimeUnix: t.Unix(),
	}
}

// GetInstalledPackages returns the list of installed packages
// based on the log files. It populates Name, Version, Manager and InstallTimeUnix
func GetInstalledPackages() ([]*models.Package, error) {
	var openErr error
	out := make([]*models.Package, 0)
	d, err := filepath.Abs(logDir)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(d)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), logFilePrefix) {
			// treat that file
			// d is a constant
			// entry is a file of d
			file := filepath.Join(d, entry.Name())
			f, err := os.Open(file) // #nosec G304 -- False positive: 'file' has the following format: /var/log/dpkg.log*
			if err != nil {
				openErr = errors.Join(openErr, err)
				// keep on
				continue
			}
			scanner := bufio.NewScanner(f)
			// scan the lines of the file
			for scanner.Scan() {
				if pkg := parseLogLine(scanner.Text()); pkg != nil {
					// add package to machine
					out = append(out, pkg)
				}
			}
			if f.Close() != nil {
				// it returns an error if the f has already been closed
				continue
			}
		}
	}
	return out, openErr
}
