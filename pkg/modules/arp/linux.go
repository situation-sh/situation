//go:build linux

package arp

import (
	"fmt"
	"net"

	"github.com/situation-sh/situation/pkg/utils"
	"github.com/vishvananda/netlink"
)

// neighMAC returns a copy of the MAC address
func neighMAC(n netlink.Neigh) net.HardwareAddr {
	length := len(n.HardwareAddr)
	mac := make(net.HardwareAddr, length)
	copy(mac, n.HardwareAddr)
	return mac
}

// neighIP returns a copy of the IP address
func neighIP(n netlink.Neigh) net.IP {
	length := len(n.IP)
	ip := make(net.IP, length)
	copy(ip, n.IP)
	return ip
}

func neighToARPEntry(n netlink.Neigh) ARPEntry {
	return ARPEntry{
		Family:         n.Family,
		InterfaceIndex: n.LinkIndex,
		MAC:            neighMAC(n),
		IP:             neighIP(n),
		State:          LinuxState(n.State),
		VLAN:           n.Vlan,
	}
}

func GetARPTable() ([]ARPEntry, error) {
	entries := make([]ARPEntry, 0)

	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	for _, link := range links {
		attr := link.Attrs()
		neighs, err := netlink.NeighList(attr.Index, 0)
		if err != nil {
			// just print the error and continue
			fmt.Println(err)
			continue
		}
		for _, neigh := range neighs {
			entry := neighToARPEntry(neigh)

			// ignore 0 and 255 in case of IPv4
			if utils.IsReserved(entry.IP) {
				continue
			}

			// ignore only incomplete. There is a problem with the states
			// output by linux. They do not seem to be the same as windows
			// ones (see state.go file).
			if entry.State != Reachable && entry.State != Stale && entry.State != Delay && entry.State != Probe {
				continue
			}

			if entry.IP.IsGlobalUnicast() {
				entries = append(entries, entry)
			}
		}
	}
	return entries, nil
}
