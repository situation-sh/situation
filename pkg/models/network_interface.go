package models

import (
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/uptrace/bun"
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

func NewNetworkInterfaceFlags(flags net.Flags) NetworkInterfaceFlags {
	return NetworkInterfaceFlags{
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
	bun.BaseModel `bun:"table:network_interfaces"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	Name      string                `bun:"name,nullzero,unique:machine_nic_name" json:"name,omitempty" jsonschema:"description=name of the network interface,example=Ethernet,example=eno1,example=eth0"`
	MAC       string                `bun:"mac,nullzero,unique:machine_mac_tag" json:"mac,omitempty" jsonschema:"description=L2 MAC address of the interface,example=74:79:27:ea:55:d3,example=93:83:e4:15:39:b2,pattern=^([A-F0-9]{2}:){5}[A-F0-9]{2}$"`
	MACVendor string                `bun:"mac_vendor,nullzero" json:"mac_vendor,omitempty" jsonschema:"description=vendor of the MAC address,example=Dell,example=Cisco,Systems,example=Intel"`
	IP        []string              `bun:"ip,nullzero,array" json:"ip,omitempty" jsonschema:"description=IPv4 address of the interface (single IP assumed),type=string,format=ipv4,example=192.168.8.1,example=10.0.0.17"`
	Gateway   string                `bun:"gateway,nullzero" json:"gateway,omitempty" jsonschema:"description=Gateway IPv4 address (main outgoing endpoint),type=string,format=ipv4,example=192.168.0.1,example=10.0.0.1"`
	Flags     NetworkInterfaceFlags `bun:"flags,type:json" json:"flags,omitempty" jsonschema:"description=Network interface flags"`
	Tag       string                `bun:"tag,unique:machine_mac_tag" json:"tag,omitempty" jsonschema:"description=Extra tag to identify the network interface,example=management,example=internal"`

	// Belongs-to relationship
	MachineID int64    `bun:"machine_id,unique:machine_nic_name,nullzero,unique:machine_mac_tag" json:"machine_id,omitempty" jsonschema:"description=ID of the machine this network interface belongs to"`
	Machine   *Machine `bun:"rel:belongs-to,join:machine_id=id"`

	// Has-many relationship
	Subnetworks []*Subnetwork `bun:"m2m:network_interface_subnets,join:NetworkInterface=Subnetwork"`

	// Many-to-many relationship via ApplicationEndpoint join table
	Applications []*Application `bun:"m2m:application_endpoints,join:NetworkInterface=Application" json:"applications,omitempty" jsonschema:"description=list of applications associated with this network interface"`
}

// networkInterfaceUnmarshallingAlias is a mirror of NetworkInterface
// for unmarshalling purpose. Golang does not allows to easily unmarshal
// net.IP or net.HardwareAddr
type networkInterfaceUnmarshallingAlias struct {
	Name    string                `json:"name,omitempty" jsonschema:"description=name of the network interface,example=Ethernet,example=eno1,example=eth0"`
	MAC     string                `json:"mac,omitempty" jsonschema:"description=L2 MAC address of the interface,example=74:79:27:ea:55:d2,example=93:83:e4:15:39:b2"`
	IP      []string              `json:"ip,omitempty" jsonschema:"description=IPv4 address of the interface (single IP assumed),type=string,format=ipv4,example=192.168.8.1,example=10.0.0.17"`
	Gateway string                `json:"gateway,omitempty" jsonschema:"description=Gateway IPv4 address (main outgoing endpoint),type=string,format=ipv4,example=192.168.0.1,example=10.0.0.1"`
	Flags   NetworkInterfaceFlags `json:"flags,omitempty" jsonschema:"description=Network interface flags"`
}

// MarshalJSON is used to customize the marshalling of the
// NetworkInterface, especially for the MAC attribute
func (nic *NetworkInterface) MarshalJSON() ([]byte, error) {
	type Alias NetworkInterface

	mac := ""
	if nic.MAC != "" {
		mac = nic.MAC
	}

	return json.Marshal(&struct {
		MAC string `json:"mac,omitempty" jsonschema:"description=L2 MAC address of the interface,example=74:79:27:ea:55:d2,example=93:83:e4:15:39:b2,pattern=^([A-F0-9]{2}:){5,7}[A-F0-9]{2}$"`
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
	nic.IP = alias.IP
	// nic.MaskSize = alias.MaskSize
	// nic.IP6 = alias.IP6
	// nic.Mask6Size = alias.Mask6Size
	nic.Gateway = alias.Gateway
	nic.Flags = alias.Flags

	// ignore parsing error (return nil)
	nic.MAC = alias.MAC

	return nil
}

// Match check if the network interface matches both the IP and MAC address
// (the match is ignored for IP if ip is nil, same for MAC)
// func (nic *NetworkInterface) Match(ip net.IP, mac net.HardwareAddr) bool {
// 	if ip == nil && mac == nil {
// 		return false
// 	}

// 	if ip != nil && !(ip.String() == nic.IP || ip.String() == nic.IP6) {
// 		return false
// 	}

// 	if mac != nil && mac.String() != nic.MAC {
// 		return false
// 	}

// 	return true
// }

// Network returns the IPv4 network attached to this nic
// func (nic *NetworkInterface) Network() *net.IPNet {
// 	if nic.IP == "" {
// 		return nil
// 	}
// 	return &net.IPNet{
// 		IP:   net.ParseIP(nic.IP),
// 		Mask: net.CIDRMask(nic.MaskSize, 32),
// 	}
// }

// Network6 returns the IPv6 network attached to this nic
// func (nic *NetworkInterface) Network6() *net.IPNet {
// 	if nic.IP6 == "" {
// 		return nil
// 	}
// 	return &net.IPNet{
// 		IP:   net.ParseIP(nic.IP6),
// 		Mask: net.CIDRMask(nic.MaskSize, 128),
// 	}
// }

// Merge update the base network interface with information from
// the second given in parameters
// func (nic *NetworkInterface) Merge(nic0 *NetworkInterface) {
// 	if nic.Name == "" {
// 		nic.Name = nic0.Name
// 	}
// 	if nic.MAC == "" {
// 		nic.MAC = nic0.MAC
// 	}
// 	if nic.IP == "" {
// 		nic.IP = nic0.IP
// 	}
// 	if nic.IP6 == "" {
// 		nic.IP6 = nic0.IP6
// 	}
// 	if nic.MaskSize <= 0 {
// 		nic.MaskSize = nic0.MaskSize
// 	}
// 	if nic.Mask6Size <= 0 {
// 		nic.Mask6Size = nic0.Mask6Size
// 	}
// 	if nic.Gateway == "" {
// 		nic.Gateway = nic0.Gateway
// 	}
// 	// if nic.Flags == nil {
// 	// copy flags
// 	nic.Flags = nic0.Flags
// 	// }
// }

func (nic *NetworkInterface) SetFlags(flags net.Flags) {
	nic.Flags = NewNetworkInterfaceFlags(flags)
}

func (NetworkInterface) JSONSchemaExtend(schema *jsonschema.Schema) {
	if macSchema, ok := schema.Properties.Get("mac"); ok {
		macSchema.Pattern = `^([a-fA-F0-9]{2}:){5,7}[a-fA-F0-9]{2}$`
	}
}

func (nic *NetworkInterface) IPs() []net.IP {
	ips := make([]net.IP, 0)
	for _, ipStr := range nic.IP {
		ipStr = strings.TrimSpace(ipStr)
		ip := net.ParseIP(ipStr)
		if ip != nil {
			ips = append(ips, ip)
		}
	}
	return ips
}
