//go:build windows
// +build windows

package arp

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func GetIpNetTable2() (MIBIpNetTable2, error) {
	// find the procedure
	proc, err := iphlpapi.FindProc("GetIpNetTable2")
	if err != nil {
		return nil, err
	}

	// find the procedure to free table
	free, err := iphlpapi.FindProc("FreeMibTable")
	if err != nil {
		return nil, err
	}

	// The GetIpNetTable2 procedure inside iphlpapi.dll
	// is responsible for memory allocation. So we merely
	// inits a (nil) pointer, that will be passed to the dll.
	var data *rawMIBIpNetTable2

	// call the procedure
	errno, _, _ := proc.Call(0, uintptr(unsafe.Pointer(&data)))
	// free the memory allocated by the dll
	defer free.Call(uintptr(unsafe.Pointer(data)))

	// Check the result
	switch syscall.Errno(errno) {
	case windows.ERROR_SUCCESS:
		err = nil
	case windows.ERROR_NOT_ENOUGH_MEMORY:
		err = fmt.Errorf("Insufficient memory resources are available to complete the operation.")
	case windows.ERROR_INVALID_PARAMETER:
		err = fmt.Errorf("An invalid parameter was passed to the function.")
	case windows.ERROR_NOT_FOUND:
		err = fmt.Errorf("No neighbor IP address entries as specified in the Family parameter were found.")
	case windows.ERROR_NOT_SUPPORTED:
		err = fmt.Errorf("The IPv4 or IPv6 transports are not configured on the local computer.")
	default:
		err = windows.GetLastError()
	}

	table := data.parse()
	return table, err
}
