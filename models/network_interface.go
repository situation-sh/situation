package models

import (
	"encoding/json"
	"net"
)

// NetworkInterface details an Ethernet/IP endpoint
type NetworkInterface struct {
	Name      string           `json:"name,omitempty"`
	MAC       net.HardwareAddr `json:"mac,omitempty"`
	IP        net.IP           `json:"ip,omitempty"`
	MaskSize  int              `json:"mask_size,omitempty"`
	IP6       net.IP           `json:"ip6,omitempty"`
	Mask6Size int              `json:"mask6_size,omitempty"`
}

// MarshalJSON is used to customize the marshalling of the
// NetworkInterface, especially for the MAC attribute
func (nic *NetworkInterface) MarshalJSON() ([]byte, error) {
	type Alias NetworkInterface

	mac := ""
	if nic.MAC != nil {
		mac = nic.MAC.String()
	}

	return json.Marshal(&struct {
		MAC string `json:"mac,omitempty"`
		*Alias
	}{
		MAC:   mac,
		Alias: (*Alias)(nic),
	})
}

// Match check if the network interface matches both the IP and MAC address
// (the match is ignored for IP if ip is nil, same for MAC)
func (nic *NetworkInterface) Match(ip net.IP, mac net.HardwareAddr) bool {
	if ip == nil && mac == nil {
		return false
	}

	if ip != nil && !(ip.Equal(nic.IP) || ip.Equal(nic.IP6)) {
		return false
	}

	if mac != nil && mac.String() != nic.MAC.String() {
		return false
	}

	return true
}

// Network returns the IPv4 network attached to this nic
func (nic *NetworkInterface) Network() *net.IPNet {
	if nic.IP == nil {
		return nil
	}
	return &net.IPNet{
		IP:   nic.IP,
		Mask: net.CIDRMask(nic.MaskSize, 32),
	}
}

// Network6 returns the IPv6 network attached to this nic
func (nic *NetworkInterface) Network6() *net.IPNet {
	if nic.IP6 == nil {
		return nil
	}
	return &net.IPNet{
		IP:   nic.IP6,
		Mask: net.CIDRMask(nic.MaskSize, 128),
	}
}
