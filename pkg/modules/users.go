// LINUX(LocalUsersModule) ok
// WINDOWS(LocalUsersModule) ok
// MACOS(LocalUsersModule) no
// ROOT(LocalUsersModule) no
package modules

import (
	"context"
	"fmt"

	"github.com/situation-sh/situation/pkg/modules/localusers"
)

func init() {
	registerModule(&LocalUsersModule{})
}

// LocalUsersModule reads package information from the dpkg package manager.
//
// This module is relevant for distros that use dpkg, like debian, ubuntu and their
// derivatives. It only uses the standard library.
//
// It reads `/var/log/dpkg.log` and also files from `/var/lib/dpkg/info/`.
type LocalUsersModule struct {
	BaseModule
}

func (m *LocalUsersModule) Name() string {
	return "users"
}

func (m *LocalUsersModule) Dependencies() []string {
	// host-basic is to check the distribution
	// netstat is to only fill the packages that have a running app
	// (see models.Machine.InsertPackages)
	return []string{"host-basic"}
}

func (m *LocalUsersModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	hostID := storage.GetHostID(ctx)
	if hostID == 0 {
		return fmt.Errorf("no host found in storage")
	}

	users, err := localusers.ListUsers()
	if err != nil {
		return err
	}

	for _, u := range users {
		u.MachineID = hostID
		logger.
			WithField("username", u.Username).
			WithField("uid", u.UID).
			WithField("gid", u.GID).
			WithField("domain", u.Domain).
			Debug("User found")
	}

	// insert all users
	_, err = storage.DB().
		NewInsert().
		Model(&users).
		On("CONFLICT (machine_id, uid) DO UPDATE").
		Set("updated_at = CURRENT_TIMESTAMP").
		Exec(ctx)
	return err
}
