//go:build linux

package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strings"
)

// GetUser returns the user running the process with the given PID.
// It reads /proc/[pid]/status to get the UID and uses os/user to lookup user details.
func GetProcessUser(pid int) (*user.User, error) {
	if pid <= 0 {
		return nil, fmt.Errorf("the PID is not strictly positive")
	}

	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	if !FileExists(statusPath) {
		return nil, fmt.Errorf("cannot retrieve status file for process with pid=%d", pid)
	}

	file, err := os.Open(statusPath) // #nosec G304 -- False positive
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var uid, gid string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Uid line format: "Uid:\treal\teffective\tsaved\tfilesystem"
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				uid = fields[1] // real UID
			}
		}
		// Gid line format: "Gid:\treal\teffective\tsaved\tfilesystem"
		if strings.HasPrefix(line, "Gid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				gid = fields[1] // real GID
			}
		}
		if uid != "" && gid != "" {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if uid == "" {
		return nil, fmt.Errorf("could not find UID for process with pid=%d", pid)
	}

	return user.LookupId(uid)
}
