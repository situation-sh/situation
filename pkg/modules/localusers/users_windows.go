//go:build windows

package localusers

import (
	"os/user"
	"syscall"
	"unsafe"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/utils"
	"golang.org/x/sys/windows"
)

var (
	modNetapi32     = windows.NewLazySystemDLL("netapi32.dll")
	procNetUserEnum = modNetapi32.NewProc("NetUserEnum")
	procNetApiFree  = modNetapi32.NewProc("NetApiBufferFree")
)

const (
	MAX_PREFERRED_LENGTH  = 0xFFFFFFFF
	FILTER_NORMAL_ACCOUNT = 0x0002
	ERROR_MORE_DATA       = 234
)

type USER_INFO_0 struct {
	Name *uint16
}

func netApiBufferFree(buf uintptr) {
	_, _, _ = procNetApiFree.Call(buf)
}

func enumLocalUsernames() ([]string, error) {
	var (
		bufPtr       uintptr
		entriesRead  uint32
		totalEntries uint32
		resumeHandle uint32
		out          []string
	)

	for {
		r1, _, _ := procNetUserEnum.Call(
			0,          // servername NULL => local machine
			uintptr(0), // level 0 => USER_INFO_0
			uintptr(FILTER_NORMAL_ACCOUNT),
			uintptr(unsafe.Pointer(&bufPtr)),
			uintptr(MAX_PREFERRED_LENGTH),
			uintptr(unsafe.Pointer(&entriesRead)),
			uintptr(unsafe.Pointer(&totalEntries)),
			uintptr(unsafe.Pointer(&resumeHandle)),
		)

		if r1 != 0 && r1 != ERROR_MORE_DATA {
			return nil, syscall.Errno(r1)
		}

		if bufPtr != 0 {
			arr := unsafe.Slice((*USER_INFO_0)(unsafe.Pointer(bufPtr)), entriesRead)
			for _, u := range arr {
				out = append(out, windows.UTF16PtrToString(u.Name))
			}
			netApiBufferFree(bufPtr)
			bufPtr = 0
		}

		if r1 != ERROR_MORE_DATA {
			break
		}
	}

	return out, nil
}

func ListUsers() ([]*models.User, error) {
	users := make([]*models.User, 0)
	names, err := enumLocalUsernames()
	if err != nil {
		panic(err)
	}

	for _, username := range names {
		user, err := user.Lookup(username)
		if err != nil {
			continue
		}

		users = append(users, &models.User{
			UID:      user.Uid,
			GID:      user.Gid,
			Name:     user.Name,
			Username: user.Username,
		})

		// assign domain
		sid, err := windows.StringToSid(user.Uid)
		if err == nil {
			domain, _, err := utils.LookupAccountSid(sid)
			if err == nil {
				users[len(users)-1].Domain = domain
			}
		}
		// domain, _, err := domainFromSID(user.Uid)
		// if err == nil {
		// 	users[len(users)-1].Domain = domain
		// }

	}
	return users, nil
}
