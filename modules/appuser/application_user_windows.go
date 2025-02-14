//go:build windows
// +build windows

package appuser

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/situation-sh/situation/models"
)

var (
	advapi32                = syscall.NewLazyDLL("advapi32.dll")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess         = kernel32.NewProc("OpenProcess")
	procCloseHandle         = kernel32.NewProc("CloseHandle")
	procOpenProcessToken    = advapi32.NewProc("OpenProcessToken")
	procGetTokenInformation = advapi32.NewProc("GetTokenInformation")
	procLookupAccountSidW   = advapi32.NewProc("LookupAccountSidW")
)

const (
	PROCESS_QUERY_INFORMATION = 0x0400
	TOKEN_QUERY               = 0x0008
	TokenUser                 = 1
)

type SIDAndAttributes struct {
	Sid        *syscall.SID
	Attributes uint32
}

func PopulateApplication(app *models.Application) error {
	if app.PID == 0 {
		return fmt.Errorf("Nul PID as input")
	}
	user := models.WindowsUser{}

	// Open the process
	hProcess, _, err := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_INFORMATION),
		0, // No inheritance
		uintptr(app.PID),
	)
	if hProcess == 0 {
		return fmt.Errorf("failed to open process %d: %v", app.PID, err)
	}
	defer procCloseHandle.Call(hProcess)

	// Open the process token
	var hToken syscall.Handle
	r, _, err := procOpenProcessToken.Call(
		hProcess,
		uintptr(TOKEN_QUERY),
		uintptr(unsafe.Pointer(&hToken)),
	)
	if r == 0 {
		return fmt.Errorf("failed to open process token: %v", err)
	}
	defer procCloseHandle.Call(uintptr(hToken))

	// Get token information size
	var tokenInfoLength uint32
	r, _, err = procGetTokenInformation.Call(
		uintptr(hToken),
		uintptr(TokenUser),
		0,
		0,
		uintptr(unsafe.Pointer(&tokenInfoLength)),
	)
	if r == 0 && syscall.Errno(err.(syscall.Errno)) != syscall.ERROR_INSUFFICIENT_BUFFER {
		return fmt.Errorf("failed to get token information size: %v", err)
	}

	// Allocate memory for token information
	tokenInfo := make([]byte, tokenInfoLength)
	r, _, err = procGetTokenInformation.Call(
		uintptr(hToken),
		uintptr(TokenUser),
		uintptr(unsafe.Pointer(&tokenInfo[0])),
		uintptr(tokenInfoLength),
		uintptr(unsafe.Pointer(&tokenInfoLength)),
	)
	if r == 0 {
		return fmt.Errorf("failed to get token information: %v", err)
	}

	// Extract the SID from token information
	tokenUser := (*SIDAndAttributes)(unsafe.Pointer(&tokenInfo[0]))
	sid := tokenUser.Sid

	// Lookup the account name
	var name [256]uint16
	var domain [256]uint16
	nameLen := uint32(len(name))
	domainLen := uint32(len(domain))
	var sidType uint32

	r, _, err = procLookupAccountSidW.Call(
		0,
		uintptr(unsafe.Pointer(sid)),
		uintptr(unsafe.Pointer(&name[0])),
		uintptr(unsafe.Pointer(&nameLen)),
		uintptr(unsafe.Pointer(&domain[0])),
		uintptr(unsafe.Pointer(&domainLen)),
		uintptr(unsafe.Pointer(&sidType)),
	)
	if r == 0 {
		return fmt.Errorf("failed to lookup account SID: %v", err)
	}

	// Convert UTF-16 strings to Go strings
	user.Domain = syscall.UTF16ToString(domain[:domainLen])
	user.Username = syscall.UTF16ToString(name[:nameLen])
	user.SID, err = sid.String()
	if err != nil {
		return fmt.Errorf("error while converting SID to string: %w", err)
	}

	// populate the application
	app.User = &user
	return nil
}
