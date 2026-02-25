//go:build windows

package utils

import (
	"fmt"
	"os/user"

	"golang.org/x/sys/windows"
)

// GetProcessUser returns the user running the process with the given PID.
// It opens the process, retrieves its token, and looks up the token owner.
func GetProcessUser(pid int) (*user.User, error) {
	if pid <= 0 {
		return nil, fmt.Errorf("the PID is not strictly positive")
	}

	// Open the process with query information access
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		// Try with limited access rights
		handle, err = windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
		if err != nil {
			return nil, fmt.Errorf("failed to open process %d: %w", pid, err)
		}
	}
	defer windows.CloseHandle(handle)

	// Open the process token
	var token windows.Token
	err = windows.OpenProcessToken(handle, windows.TOKEN_QUERY, &token)
	if err != nil {
		return nil, fmt.Errorf("failed to open process token for pid=%d: %w", pid, err)
	}
	defer token.Close()

	// Get the token user
	tokenUser, err := token.GetTokenUser()
	if err != nil {
		return nil, fmt.Errorf("failed to get token user for pid=%d: %w", pid, err)
	}

	// Convert SID to string
	sidStr := tokenUser.User.Sid.String()

	// Lookup account name and domain from SID
	_, username, err := LookupAccountSid(tokenUser.User.Sid)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup account for SID %s: %w", sidStr, err)
	}

	// Try to get more user info via os/user
	return user.Lookup(username)
}

// LookupAccountSid returns the domain and account name for a SID.
func LookupAccountSid(sid *windows.SID) (domain, name string, err error) {
	var nameLen uint32
	var domainLen uint32
	var use uint32

	// Probe sizes
	// If the function fails because the buffer is too small or if cchName is zero,
	// cchName receives the required buffer size, including the terminating null character.
	//
	// BOOL LookupAccountSidA(
	//   [in, optional]  LPCSTR        lpSystemName,
	//   [in]            PSID          Sid,
	//   [out, optional] LPSTR         Name,
	//   [in, out]       LPDWORD       cchName,
	//   [out, optional] LPSTR         ReferencedDomainName,
	//   [in, out]       LPDWORD       cchReferencedDomainName,
	//   [out]           PSID_NAME_USE peUse
	// );
	_ = windows.LookupAccountSid(
		nil,
		sid,
		nil,
		&nameLen,
		nil,
		&domainLen,
		&use,
	)

	if nameLen == 0 {
		return "", "", fmt.Errorf("failed to get buffer sizes for SID lookup")
	}

	nameBuf := make([]uint16, nameLen)
	domainBuf := make([]uint16, domainLen)

	err = windows.LookupAccountSid(
		nil,
		sid,
		&nameBuf[0], &nameLen,
		&domainBuf[0], &domainLen,
		&use,
	)
	if err != nil {
		return "", "", err
	}

	return windows.UTF16ToString(domainBuf), windows.UTF16ToString(nameBuf), nil
}
