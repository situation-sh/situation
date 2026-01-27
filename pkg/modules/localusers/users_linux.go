//go:build linux

package localusers

import (
	"bufio"
	"os"
	"os/user"
	"strings"

	"github.com/situation-sh/situation/pkg/models"
)

const USER_SOURCE = "/etc/passwd"

func ListUsers() ([]*models.User, error) {
	users := make([]*models.User, 0)
	file, err := os.Open(USER_SOURCE)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 3 {
			continue
		}
		uid := parts[2]
		if u, err := user.LookupId(uid); err == nil {
			users = append(users, &models.User{
				UID:      u.Uid,
				Name:     u.Name,
				Username: u.Username,
				GID:      u.Gid,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
