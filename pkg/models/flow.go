package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Flow struct {
	bun.BaseModel `bun:"table:flows"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	SrcApplicationID int64        `bun:"src_application_id,nullzero,unique:flow_src_dst"`
	SrcApplication   *Application `bun:"rel:belongs-to,join:src_application_id=id"`

	SrcNetworkInterfaceID int64             `bun:"src_network_interface_id,nullzero"`
	SrcNetworkInterface   *NetworkInterface `bun:"rel:belongs-to,join:src_network_interface_id=id"`

	SrcAddr string `bun:"src_addr,unique:flow_src_dst" json:"src_addr" jsonschema:"description=source IP address of the flow,example=192.168.0.1,example=fe80::12b8:43e7:11f0:a406"`

	DstEndpointID int64                `bun:"dst_endpoint_id,nullzero,unique:flow_src_dst"`
	DstEndpoint   *ApplicationEndpoint `bun:"rel:belongs-to,join:dst_endpoint_id=id"`
}
