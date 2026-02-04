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

// LocalUsersModule lists all local user accounts on the system.
//
// On **Linux**, the module reads `/etc/passwd` to enumerate user entries.
// Each UID is then resolved through the standard `os/user` library to
// retrieve the full user details.
//
// On **Windows**, the module calls the Win32 `NetUserEnum` API
// (from `netapi32.dll`) to enumerate local accounts filtered to normal
// user accounts. Each username is then resolved with `os/user.Lookup`,
// and the user's domain is determined by converting the SID via
// `LookupAccountSid`.
//
// The collected users are stored in the database with an upsert strategy
// based on `(machine_id, uid)`.
type LocalUsersModule struct {
	BaseModule
}

func (m *LocalUsersModule) Name() string {
	return "local-users"
}

func (m *LocalUsersModule) Dependencies() []string {
	// host-basic is to check the distribution
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
