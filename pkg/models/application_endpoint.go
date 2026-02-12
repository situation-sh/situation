package models

import (
	"time"

	"github.com/invopop/jsonschema"
	"github.com/uptrace/bun"
)

type TLS struct {
	Subject            string    `json:"subject,omitempty" jsonschema:"description=subject of the certificate,example=www.example.com,example=CN=www.example.com,O=Example Inc.,C=US"`
	Issuer             string    `json:"issuer,omitempty" jsonschema:"description=issuer of the certificate,example=Let's Encrypt Authority X3,O=Let's Encrypt,C=US"`
	NotBefore          time.Time `json:"not_before,omitempty" jsonschema:"description=UNIX timestamp of the certificate not before date,example=1670520587"`
	NotAfter           time.Time `json:"not_after,omitempty" jsonschema:"description=UNIX timestamp of the certificate not after date,example=1670520587"`
	SerialNumber       string    `json:"serial_number,omitempty" jsonschema:"description=serial number of the certificate,example=1234567890"`
	SignatureAlgorithm string    `json:"signature_algorithm,omitempty" jsonschema:"description=signature algorithm used to sign the certificate,example=SHA256withRSA"`
	PublicKeyAlgorithm string    `json:"public_key_algorithm,omitempty" jsonschema:"description=public key algorithm used in the certificate,example=RSA,example=ECDSA"`
	// Signature          string    `json:"signature,omitempty" jsonschema:"description=base64 encoded signature of the certificate"`
	SHA1Fingerprint   string   `json:"sha1_fingerprint,omitempty" jsonschema:"description=SHA1 fingerprint of the certificate,example=3A:5B:7C:8D:9E:0F:1A:2B:3C:4D:5E:6F:7A:8B:9C"`
	SHA256Fingerprint string   `json:"sha256_fingerprint,omitempty" jsonschema:"description=SHA256 fingerprint of the certificate,example=3A:5B:7C:8D:9E:0F:1A:2B:3C:4D"`
	DNSNames          []string `json:"dns_names,omitempty" jsonschema:"description=list of DNS names included in the certificate,example=[\"www.example.com\",\"example.com\"]"`
}

type JA4 struct {
	JA4  string `json:"ja4,omitempty" jsonschema:"description=JA4 TLS client fingerprint,example=t13d1516h2_8daaf6152771_02713d6af862"`
	JA4S string `json:"ja4s,omitempty" jsonschema:"description=JA4S TLS server fingerprint,example=t130200_1302_a56c5b993250 "`
	JA4X string `json:"ja4x,omitempty" jsonschema:"description=JA4X TLS cert fingerprint,example=2bab15409345_af684594efb4_000000000000"`
}

type Fingerprints struct {
	JA4 *JA4 `json:"ja4,omitempty" jsonschema:"description=JA4 fingerprints"`
}

// ApplicationEndpoint is a structure used by Application
// to tell that the app listens on given addr and port
type ApplicationEndpoint struct {
	bun.BaseModel `bun:"table:application_endpoints"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	Port                 uint16        `bun:"port,type:integer,unique:port_protocol_addr_network_interface_id" json:"port" jsonschema:"description=port number,example=22,example=80,example=443,minimum=1,maximum=65535"`
	Protocol             string        `bun:"protocol,unique:port_protocol_addr_network_interface_id" json:"protocol" jsonschema:"description=transport layer protocol,example=tcp,example=udp"`
	Addr                 string        `bun:"addr,unique:port_protocol_addr_network_interface_id" json:"addr" jsonschema:"description=IP address the application listens on,example=127.0.0.1,example=fe80::12b8:43e7:11f0:a406"`
	TLS                  *TLS          `bun:"tls,type:json" json:"tls,omitempty" jsonschema:"description=TLS information if the endpoint is using TLS"`
	Fingerprints         *Fingerprints `bun:"fingerprints,type:json" json:"fingerprints,omitempty" jsonschema:"description=application fingerprints"`
	ApplicationProtocols []string      `bun:"application_protocols,nullzero" json:"application_protocols,omitempty" jsonschema:"description=list of application layer protocols detected on this endpoint,example=[\"http\",\"http/2\"]"`

	SaaS string `bun:"saas,nullzero" json:"saas,omitempty" jsonschema:"description=if the application is identified as a SaaS, the name of the SaaS,example=GitHub,example=Google Workspace"`

	ApplicationID int64        `bun:"application_id,nullzero"`
	Application   *Application `bun:"rel:belongs-to,join:application_id=id"`

	NetworkInterfaceID int64             `bun:"network_interface_id,unique:port_protocol_addr_network_interface_id"` // can be null
	NetworkInterface   *NetworkInterface `bun:"rel:belongs-to,join:network_interface_id=id"`

	// Has-many relationship
	IncomingFlows []*Flow `bun:"rel:has-many,join:id=dst_endpoint_id"`
}

func (ApplicationEndpoint) JSONSchemaExtend(schema *jsonschema.Schema) {
	if addrSchema, ok := schema.Properties.Get("addr"); ok {
		addrSchema.AnyOf = []*jsonschema.Schema{
			{Type: "string", Format: "ipv4"},
			{Type: "string", Format: "ipv6"},
		}
	}
}

// EndpointPolicy defines policies applied to an ApplicationEndpoint
// such as filtering or forwarding rules
type EndpointPolicy struct {
	bun.BaseModel `bun:"table:endpoint_policies"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	// The endpoint this policy applies to
	EndpointID int64                `bun:"endpoint_id,notnull,unique:endpoint_action_src"`
	Endpoint   *ApplicationEndpoint `bun:"rel:belongs-to,join:endpoint_id=id"`

	// What happens to the traffic
	// "accept", "drop", "reject", "forward"
	Action string `bun:"action,unique:endpoint_action_src" json:"action"`

	// Where the traffic comes from — use one or the other:
	// For forwarding: the upstream endpoint (host:8080 -> container:80)
	SrcEndpointID int64                `bun:"src_endpoint_id,nullzero,unique:endpoint_action_src"`
	SrcEndpoint   *ApplicationEndpoint `bun:"rel:belongs-to,join:src_endpoint_id=id"`
	// For filtering: a network source (CIDR), empty = any
	SrcAddr string `bun:"src_addr,nullzero,unique:endpoint_action_src" json:"src_addr,omitempty"`

	// Rule ordering (lower = matched first, matters for accept+drop combos)
	Priority int `bun:"priority" json:"priority"`

	// How this was discovered
	Source string `bun:"source" json:"source"` // "docker", "iptables", "nftables", "ufw"
}
