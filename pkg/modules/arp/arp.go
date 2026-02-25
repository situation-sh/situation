// Package arp provides a general interface to retrieve the ARP table
// on both Linux and Windows
package arp

import (
	"net"
)

type ARPEntry struct {
	Family         int
	InterfaceIndex int
	MAC            net.HardwareAddr
	IP             net.IP
	State          ARPEntryState
	VLAN           int
}

// FilterARPTableByNetwork returns the ARP entries that belong to
// a given network
func FilterARPTableByNetwork(table []ARPEntry, network *net.IPNet) []ARPEntry {
	out := make([]ARPEntry, 0)
	for _, record := range table {
		if network.Contains(record.IP) {
			out = append(out, record)
		}
	}
	return out
}
