//go:build linux
// +build linux

package appuser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/situation-sh/situation/models"
)

func PopulateApplication(app *models.Application) error {
	if app.PID == 0 {
		return fmt.Errorf("Nul PID as input")
	}
	user, err := getLinuxUser(app.PID)
	if err != nil {
		return err
	}

	// populate the app
	app.User = user
	return nil
}

func getLinuxUser(pid uint) (*models.LinuxUser, error) {
	path := fmt.Sprintf("/proc/%d/status", pid)
	file, err := os.Open(path) // #nosec G304 -- input is controlled
	if err != nil {
		return nil, err
	}

	var uids, gids []uint
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "Uid:":
			uids = parseIDs(fields[1:])
		case "Gid:":
			gids = parseIDs(fields[1:])
		}
	}

	if len(uids) < 4 || len(gids) < 4 {
		return nil, errors.New("failed to parse UID/GID data")
	}

	return &models.LinuxUser{
		UID:   getLinuxID(uids[0], true),
		EUID:  getLinuxID(uids[1], true),
		SUID:  getLinuxID(uids[2], true),
		FSUID: getLinuxID(uids[3], true),
		GID:   getLinuxID(gids[0], false),
		EGID:  getLinuxID(gids[1], false),
		SGID:  getLinuxID(gids[2], false),
		FSGID: getLinuxID(gids[3], false),
	}, file.Close()
}

func parseIDs(fields []string) []uint {
	var ids []uint
	for _, field := range fields {
		if id, err := strconv.Atoi(field); err == nil {
			ids = append(ids, uint(id))
		}
	}
	return ids
}

func getLinuxID(id uint, isUser bool) *models.LinuxID {
	var name string
	if isUser {
		usr, err := user.LookupId(strconv.Itoa(int(id)))
		if err == nil {
			name = usr.Username
		}
	} else {
		grp, err := user.LookupGroupId(strconv.Itoa(int(id)))
		if err == nil {
			name = grp.Name
		}
	}
	return &models.LinuxID{ID: id, Name: name}
}
