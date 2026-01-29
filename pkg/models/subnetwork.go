package models

import (
	"net"
	"time"

	"github.com/uptrace/bun"
)

type NetworkInterfaceSubnet struct {
	bun.BaseModel `bun:"table:network_interface_subnets"`

	NetworkInterfaceID int64             `bun:"network_interface_id,pk"`
	NetworkInterface   *NetworkInterface `bun:"rel:belongs-to,join:network_interface_id=id"`

	SubnetworkID int64       `bun:"subnetwork_id,pk"`
	Subnetwork   *Subnetwork `bun:"rel:belongs-to,join:subnetwork_id=id"`

	// dummy columns for constraints only
	MACSubnet string `bun:"mac_cidr,unique,notnull,nullzero"`
}

type Subnetwork struct {
	bun.BaseModel `bun:"table:subnetworks"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	NetworkCIDR string `bun:"network_cidr,unique,notnull" json:"network_cidr" jsonschema:"description=network in CIDR notation,example=192.168.1.0/24"`
	NetworkAddr string `bun:"network_addr,type:text,notnull" json:"network_addr,omitempty" jsonschema:"description=network address,example=192.168.1.0"`
	MaskSize    int    `bun:"mask_size,notnull" json:"mask_size,omitempty" jsonschema:"description=subnetwork mask size,example=24,minimum=0,maximum=32"`
	IPVersion   int    `bun:"ip_version,notnull" json:"ip_version,omitempty" jsonschema:"description=IP version,example=4,minimum=4,maximum=6"`
	Gateway     string `bun:"gateway,type:text" json:"gateway,omitempty" jsonschema:"description=gateway IP address,example=192.168.1.1"`
	VLANID      int    `bun:"vlan_id" json:"vlan_id,omitempty" jsonschema:"description=VLAN identifier,example=100"`

	// Has-many relationship
	NetworkInterfaces []*NetworkInterface `bun:"m2m:network_interface_subnets,join:Subnetwork=NetworkInterface"`
}

func (s *Subnetwork) IPNet() (*net.IPNet, error) {
	_, ipnet, err := net.ParseCIDR(s.NetworkCIDR)
	if err != nil {
		return nil, err
	}
	return ipnet, nil
}
