package models

import (
	"encoding/json"
	"net"
)

// NetworkInterfaceFlags give details about a network interface
// see https://pkg.go.dev/net#Flags
type NetworkInterfaceFlags struct {
	Up           bool `json:"up" jsonschema:"description=interface is administratively up"`
	Broadcast    bool `json:"broadcast,omitempty" jsonschema:"description=interface supports broadcast access capability"`
	Loopback     bool `json:"loopback,omitempty" jsonschema:"description=interface is a loopback interface"`
	PointToPoint bool `json:"point_to_point,omitempty" jsonschema:"description=interface belongs to a point-to-point link"`
	Multicast    bool `json:"multicast,omitempty" jsonschema:"description=interface supports multicast access capability"`
	Running      bool `json:"running,omitempty" jsonschema:"description=interface is in running state"`
}

func NewNetworkInterfaceFlags(flags net.Flags) *NetworkInterfaceFlags {
	return &NetworkInterfaceFlags{
		Up:           (flags & net.FlagUp) > 0,
		Broadcast:    (flags & net.FlagBroadcast) > 0,
		Loopback:     (flags & net.FlagLoopback) > 0,
		PointToPoint: (flags & net.FlagPointToPoint) > 0,
		Multicast:    (flags & net.FlagMulticast) > 0,
		Running:      (flags & net.FlagRunning) > 0,
	}
}

// NetworkInterface details an Ethernet/IP endpoint
type NetworkInterface struct {
	Name      string                 `json:"name,omitempty" jsonschema:"description=name of the network interface,example=Ethernet,example=eno1,example=eth0"`
	MAC       net.HardwareAddr       `json:"mac,omitempty" jsonschema:"description=L2 MAC address of the interface,example=74:79:27:ea:55:d2,example=93:83:e4:15:39:b2"`
	IP        net.IP                 `json:"ip,omitempty" jsonschema:"description=IPv4 address of the interface (single IP assumed),type=string,format=ipv4,example=192.168.8.1,example=10.0.0.17"`
	MaskSize  int                    `json:"mask_size,omitempty" jsonschema:"description=IPv4 subnetwork mask size,example=24,example=16,minimum=0,maximum=32"`
	IP6       net.IP                 `json:"ip6,omitempty" jsonschema:"description=IPv6 address of the interface (single IP assumed),type=string,format=ipv6,example=fe80::14a:7687:d7bd:f461,example=fe80::13d4:43e1:11e0:3906"`
	Mask6Size int                    `json:"mask6_size,omitempty" jsonschema:"description=IPv6 subnetwork mask size,example=64,minimum=0,maximum=128"`
	Gateway   net.IP                 `json:"gateway,omitempty" jsonschema:"description=Gateway IPv4 address (main outgoing endpoint),type=string,format=ipv4,example=192.168.0.1,example=10.0.0.1"`
	Flags     *NetworkInterfaceFlags `json:"flags,omitempty" jsonschema:"description=Network interface flags"`
}

// networkInterfaceUnmarshallingAlias is a mirror of NetworkInterface
// for unmarshalling purpose. Golang does not allows to easily unmarshal
// net.IP or net.HardwareAddr
type networkInterfaceUnmarshallingAlias struct {
	Name      string                 `json:"name,omitempty" jsonschema:"description=name of the network interface,example=Ethernet,example=eno1,example=eth0"`
	MAC       string                 `json:"mac,omitempty" jsonschema:"description=L2 MAC address of the interface,example=74:79:27:ea:55:d2,example=93:83:e4:15:39:b2"`
	IP        string                 `json:"ip,omitempty" jsonschema:"description=IPv4 address of the interface (single IP assumed),type=string,format=ipv4,example=192.168.8.1,example=10.0.0.17"`
	MaskSize  int                    `json:"mask_size,omitempty" jsonschema:"description=IPv4 subnetwork mask size,example=24,example=16,minimum=0,maximum=32"`
	IP6       string                 `json:"ip6,omitempty" jsonschema:"description=IPv6 address of the interface (single IP assumed),type=string,format=ipv6,example=fe80::14a:7687:d7bd:f461,example=fe80::13d4:43e1:11e0:3906"`
	Mask6Size int                    `json:"mask6_size,omitempty" jsonschema:"description=IPv6 subnetwork mask size,example=64,minimum=0,maximum=128"`
	Gateway   string                 `json:"gateway,omitempty" jsonschema:"description=Gateway IPv4 address (main outgoing endpoint),type=string,format=ipv4,example=192.168.0.1,example=10.0.0.1"`
	Flags     *NetworkInterfaceFlags `json:"flags,omitempty" jsonschema:"description=Network interface flags"`
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
		MAC string `json:"mac,omitempty" jsonschema:"description=L2 MAC address of the interface,example=74:79:27:ea:55:d2,example=93:83:e4:15:39:b2,pattern=^([A-F0-9]{2}:){5}[A-F0-9]{2}$"`
		*Alias
	}{
		MAC:   mac,
		Alias: (*Alias)(nic),
	})
}

// UnmarshalJSON is an helper for the golang ecosystem (we can reuse the models
// in other golang apps)
func (nic *NetworkInterface) UnmarshalJSON(data []byte) error {
	var err error

	var alias networkInterfaceUnmarshallingAlias
	if err = json.Unmarshal(data, &alias); err != nil {
		return err
	}

	nic.Name = alias.Name
	nic.IP = net.ParseIP(alias.IP)
	nic.MaskSize = alias.MaskSize
	nic.IP6 = net.ParseIP(alias.IP6)
	nic.Mask6Size = alias.Mask6Size
	nic.Gateway = net.ParseIP(alias.Gateway)
	nic.Flags = alias.Flags

	// ignore parsing error (return nil)
	nic.MAC, err = net.ParseMAC(alias.MAC)
	if err != nil {
		nic.MAC = net.HardwareAddr{0, 0, 0, 0, 0, 0, 0, 0}
	}

	return nil
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

// Merge update the base network interface with information from
// the second given in parameters
func (nic *NetworkInterface) Merge(nic0 *NetworkInterface) {
	if nic.Name == "" {
		nic.Name = nic0.Name
	}
	if nic.MAC == nil {
		nic.MAC = nic0.MAC
	}
	if nic.IP == nil {
		nic.IP = nic0.IP
	}
	if nic.IP6 == nil {
		nic.IP6 = nic0.IP6
	}
	if nic.MaskSize <= 0 {
		nic.MaskSize = nic0.MaskSize
	}
	if nic.Mask6Size <= 0 {
		nic.Mask6Size = nic0.Mask6Size
	}
	if nic.Gateway == nil {
		nic.Gateway = nic0.Gateway
	}
	if nic.Flags == nil {
		// copy flags
		nic.Flags = nic0.Flags
	}
}

func (nic *NetworkInterface) SetFlags(flags net.Flags) {
	nic.Flags = NewNetworkInterfaceFlags(flags)
}
