package models

import (
	"time"

	"github.com/invopop/jsonschema"
	"github.com/uptrace/bun"
)

// type FlowRemoteExtra struct {
// 	TLS          *TLS          `json:"tls,omitempty" jsonschema:"description=TLS information if the flow is using TLS"`
// 	Fingerprints *Fingerprints `json:"fingerprints,omitempty" jsonschema:"description=flow fingerprints if relevant"`
// }

// Flow aims to represent a layer 4 connection
// type Flow struct {
// 	LocalAddr   net.IP           `json:"local_addr" jsonschema:"description=local IP address of the flow,example=127.0.0.1,example=fe80::12b8:43e7:11f0:a406"`
// 	LocalPort   uint16           `json:"local_port" jsonschema:"description=local port of the flow,example=22,example=443,example=49667,minimum=1,maximum=65535"`
// 	RemoteAddr  net.IP           `json:"remote_addr" jsonschema:"description=remote IP address of the flow,example=9.10.11.12,example=2607:5300:201:abcd::7c5"`
// 	RemotePort  uint16           `json:"remote_port" jsonschema:"description=remote port of the flow,example=22,example=443,example=8080,minimum=1,maximum=65535"`
// 	Protocol    string           `json:"protocol" jsonschema:"description=transport layer protocol,example=tcp,example=udp"`
// 	Status      string           `json:"status" jsonschema:"description=Link status,example=CLOSE_WAIT,example=ESTABLISHED"`
// 	RemoteExtra *FlowRemoteExtra `json:"remote_extra,omitempty" jsonschema:"description=extra information about the remote endpoint"`
// }

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

// func (f *Flow) Revert() Flow {
// 	return Flow{
// 		LocalAddr:  append(net.IP(nil), f.RemoteAddr...),
// 		LocalPort:  f.RemotePort,
// 		RemoteAddr: append(net.IP(nil), f.LocalAddr...),
// 		RemotePort: f.LocalPort,
// 		Protocol:   f.Protocol,
// 		Status:     f.Status,
// 	}
// }

// func (f *Flow) Equal(other *Flow) bool {
// 	if !f.LocalAddr.Equal(other.LocalAddr) {
// 		return false
// 	}
// 	if f.LocalPort != other.LocalPort {
// 		return false
// 	}
// 	if !f.RemoteAddr.Equal(other.RemoteAddr) {
// 		return false
// 	}
// 	if f.RemotePort != other.RemotePort {
// 		return false
// 	}
// 	if f.Protocol != other.Protocol {
// 		return false
// 	}
// 	return true
// }

func (Flow) JSONSchemaExtend(schema *jsonschema.Schema) {
	for _, prop := range []string{"local_addr", "remote_addr"} {
		if addrSchema, ok := schema.Properties.Get(prop); ok {
			addrSchema.AnyOf = []*jsonschema.Schema{
				{Type: "string", Format: "ipv4"},
				{Type: "string", Format: "ipv6"},
			}
		}
	}
}
