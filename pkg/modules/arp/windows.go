//go:build windows

package arp

import (
	"github.com/situation-sh/situation/pkg/utils"
	"golang.org/x/sys/windows"
)

var iphlpapi *windows.DLL

// defaultBufferSize is the default size of the buffer that is prepared to
// store the results of the GetIpNetTable/GetIpNetTable2 procedures.
// It is automatically increased if it is not enough
// var defaultBufferSize = 4096

func init() {
	iphlpapi = windows.MustLoadDLL("Iphlpapi.dll")
}

func GetARPTable() ([]ARPEntry, error) {
	table, err := GetIpNetTable2()
	if err != nil {
		return nil, err
	}
	entries := make([]ARPEntry, 0)
	for _, row := range table {
		entry := row.ToARPEntry()

		// ignore 0 and 255 in case of IPv4
		if utils.IsReserved(entry.IP) {
			continue
		}

		// ignore some entries
		if entry.State == Unreachable || entry.State == Incomplete {
			continue
		}
		if entry.IP.IsGlobalUnicast() {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}
