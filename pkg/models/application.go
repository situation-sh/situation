package models

import (
	"net"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/uptrace/bun"
)

// Package is a wrapper around application that stores distribution
// information of applications (executables)
type Package struct {
	bun.BaseModel `bun:"table:packages"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	Name            string `bun:"name" json:"name,omitempty" jsonschema:"description=name of the package,example=openssh,example=musl-gcc,example=python,example=texlive"`
	Version         string `bun:"version" json:"version,omitempty" jsonschema:"description=version of the package,example=8.8p1,example=1.2.3,example=3.10.8,example=2021"`
	Vendor          string `bun:"vendor" json:"vendor,omitempty" jsonschema:"description=name of the organization that produce the package or its maintainer,example=Fedora Project,example=bob@debian.org"`
	Manager         string `bun:"manager" json:"manager,omitempty" jsonschema:"description=program that manages the package installation,example=rpm,example=dpkg,example=msi,example=builtin"`
	InstallTimeUnix int64  `bun:"install_time_unix" json:"install_time,omitempty" jsonschema:"description=UNIX timestamp of the package installation,example=1670520587"`
	// Applications    []*Application `json:"applications" jsonschema:"description=list of the applications provided by this package"`
	Files []string `bun:"files,type:json" json:"-"` // ignore that field for the moment

	// Belongs-to relationship
	MachineID int64    `bun:"machine_id,notnull"`
	Machine   *Machine `bun:"rel:belongs-to,join:machine_id=id"`

	// Has-many relationship
	Applications []*Application `bun:"rel:has-many,join:id=package_id" json:"applications" jsonschema:"description=list of applications"`
}

func NewPackage() *Package {
	return &Package{
		Applications: make([]*Application, 0),
		Files:        make([]string, 0),
	}
}

// Equal check if two packages are the same. Here we assume
// that Name and Version must be set
func (pkg *Package) Equal(other *Package) bool {
	if !(pkg.Name == other.Name && len(pkg.Name) > 0) {
		return false
	}
	if !(pkg.Version == other.Version && len(pkg.Version) > 0) {
		return false
	}
	if pkg.Vendor != other.Vendor {
		return false
	}
	if pkg.Manager != other.Manager {
		return false
	}
	return true
}

// ApplicationNames return the names of the apps that are
// attached to the package
func (pkg *Package) ApplicationNames() []string {
	if len(pkg.Applications) == 0 {
		return []string{}
	}
	out := make([]string, len(pkg.Applications))
	for i, a := range pkg.Applications {
		out[i] = a.Name
	}
	return out
}

// WindowsUser is a structure that stores user info
// This is notably used to enrich the Application struct
// type WindowsUser struct {
// 	SID      string `json:"sid,omitempty" jsonschema:"description=security identifier (windows),example=S-1-5-21–7623811015-3361044348-030300820-1013"`
// 	Username string `json:"username,omitempty" jsonschema:"description=username related to the sid,example=Administrator"`
// 	Domain   string `json:"domain,omitempty" jsonschema:"description=domain related to the sid,example=DESKTOP-V1XZZQ3"`
// }

// type LinuxID struct {
// 	ID   uint   `json:"id" jsonschema:"example=1000,example=0,example=999"`
// 	Name string `json:"name,omitempty" jsonschema:"example=bob,example=root,example=docker"`
// }

// // WindowsUser is a structure that stores user info
// // This is notably used to enrich the Application struct
// type LinuxUser struct {
// 	UID   *LinuxID `json:"uid,omitempty" jsonschema:"description=the actual user who owns the process."`
// 	EUID  *LinuxID `json:"euid,omitempty" jsonschema:"description=the permissions a process has while executing"`
// 	SUID  *LinuxID `json:"suid,omitempty" jsonschema:"description=stores the EUID before a privilege change"`
// 	FSUID *LinuxID `json:"fsuid,omitempty" jsonschema:"description=file permission checks"`
// 	GID   *LinuxID `json:"gid,omitempty" jsonschema:"description=the primary group of the user who started the process."`
// 	EGID  *LinuxID `json:"egid,omitempty" jsonschema:"description=the group permissions the process has while executing"`
// 	SGID  *LinuxID `json:"sgid,omitempty" jsonschema:"description=stores the EGID before privilege changes"`
// 	FSGID *LinuxID `json:"fsgid,omitempty" jsonschema:"description=file permission checks"`
// }

// Application is a structure that represents all the
// types of apps we can have on a system
type Application struct {
	bun.BaseModel `bun:"table:applications"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	Name string   `bun:"name" json:"name,omitempty" jsonschema:"description=path (or name) of the application,example=/usr/sbin/sshd,example=/usr/bin/musl-gcc,example=C:\\Windows\\System32\\svchost.exe,example=wininit.exe,example=System"`
	Args []string `bun:"args,type:json" json:"args,omitempty" jsonschema:"description=list of arguments passed to app"` // we cannot put example right now (PR in progress: https://github.com/invopop/jsonschema/pull/31)
	// Endpoints []*ApplicationEndpoint `json:"endpoints" jsonschema:"description=list of network endpoints open by this app"`
	PID uint `bun:"pid" json:"pid,omitempty" jsonschema:"description=processus ID,example=5452,example=19420"`
	// Flows []*Flow `json:"flows" jsonschema:"description=list of flows generated by this app"`
	// User      interface{}            `json:"user,omitempty" jsonschema:"description=user running the app,oneof_ref=#/$defs/LinuxUser;#/$defs/WindowsUser"` // these definitions are manually provided
	Version  string                 `bun:"version" json:"version,omitempty"  jsonschema:"description=application version,example=9.8,example=1.2.3"`
	Protocol string                 `bun:"protocol" json:"protocol,omitempty" jsonschema:"description=protocol used to talk to the application,example=ssh,example=http"`
	Config   map[string]interface{} `bun:"config,type:json" json:"config,omitempty" jsonschema:"description=application configuration or metadata"`
	CPE      string                 `bun:"cpe" json:"cpe,omitempty" jsonschema:"description=application CPE uri,example=cpe:2.3:a:f5:nginx:*:*:*:*:*:*:*:*"`

	// Belongs-to relationship
	PackageID int64    `bun:"package_id,notnull"`
	Package   *Package `bun:"rel:belongs-to,join:package_id=id" json:"package,omitempty" jsonschema:"description=package providing this application"`
	// Many-to-many relationship
	Users []User `bun:"m2m:user_applications,join:User=Application" json:"users,omitempty" jsonschema:"description=list of users associated with this application"`
	// Many-to-many relationship via ApplicationEndpoint join table
	NetworkInterfaces []*NetworkInterface `bun:"m2m:application_endpoints,join:Application=NetworkInterface" json:"network_interfaces,omitempty" jsonschema:"description=list of network interfaces associated with this application"`
	// Has-many relationship
	Endpoints []*ApplicationEndpoint `bun:"rel:has-many,join:id=application_id" json:"endpoints,omitempty" jsonschema:"description=list of network endpoints open by this app"`
}

func NewApplication() *Application {
	return &Application{
		Endpoints: make([]*ApplicationEndpoint, 0),
		// Flows:     make([]*Flow, 0),
		Config: make(map[string]interface{}),
	}
}

type FlowRemoteExtra struct {
	TLS          *TLS          `json:"tls,omitempty" jsonschema:"description=TLS information if the flow is using TLS"`
	Fingerprints *Fingerprints `json:"fingerprints,omitempty" jsonschema:"description=flow fingerprints if relevant"`
}

// Flow aims to represent a layer 4 connection
type Flow struct {
	LocalAddr   net.IP           `json:"local_addr" jsonschema:"description=local IP address of the flow,example=127.0.0.1,example=fe80::12b8:43e7:11f0:a406"`
	LocalPort   uint16           `json:"local_port" jsonschema:"description=local port of the flow,example=22,example=443,example=49667,minimum=1,maximum=65535"`
	RemoteAddr  net.IP           `json:"remote_addr" jsonschema:"description=remote IP address of the flow,example=9.10.11.12,example=2607:5300:201:abcd::7c5"`
	RemotePort  uint16           `json:"remote_port" jsonschema:"description=remote port of the flow,example=22,example=443,example=8080,minimum=1,maximum=65535"`
	Protocol    string           `json:"protocol" jsonschema:"description=transport layer protocol,example=tcp,example=udp"`
	Status      string           `json:"status" jsonschema:"description=Link status,example=CLOSE_WAIT,example=ESTABLISHED"`
	RemoteExtra *FlowRemoteExtra `json:"remote_extra,omitempty" jsonschema:"description=extra information about the remote endpoint"`
}

func (f *Flow) Revert() Flow {
	return Flow{
		LocalAddr:  append(net.IP(nil), f.RemoteAddr...),
		LocalPort:  f.RemotePort,
		RemoteAddr: append(net.IP(nil), f.LocalAddr...),
		RemotePort: f.LocalPort,
		Protocol:   f.Protocol,
		Status:     f.Status,
	}
}

func (f *Flow) Equal(other *Flow) bool {
	if !f.LocalAddr.Equal(other.LocalAddr) {
		return false
	}
	if f.LocalPort != other.LocalPort {
		return false
	}
	if !f.RemoteAddr.Equal(other.RemoteAddr) {
		return false
	}
	if f.RemotePort != other.RemotePort {
		return false
	}
	if f.Protocol != other.Protocol {
		return false
	}
	return true
}

// type TLSSubject struct {
// 	CommonName   string `json:"common_name,omitempty" jsonschema:"description=common name of the subject,example=www.example.com"`
// 	Organization string `json:"organization,omitempty" jsonschema:"description=organization of the subject,example=Example Inc.,example=Example Ltd."`
// }

type TLS struct {
	Subject            string    `json:"subject,omitempty" jsonschema:"description=subject of the certificate,example=www.example.com,example=CN=www.example.com,O=Example Inc.,C=US"`
	Issuer             string    `json:"issuer,omitempty" jsonschema:"description=issuer of the certificate,example=Let's Encrypt Authority X3,O=Let's Encrypt,C=US"`
	NotBefore          time.Time `json:"not_before,omitempty" jsonschema:"description=UNIX timestamp of the certificate not before date,example=1670520587"`
	NotAfter           time.Time `json:"not_after,omitempty" jsonschema:"description=UNIX timestamp of the certificate not after date,example=1670520587"`
	SerialNumber       string    `json:"serial_number,omitempty" jsonschema:"description=serial number of the certificate,example=1234567890"`
	SignatureAlgorithm string    `json:"signature_algorithm,omitempty" jsonschema:"description=signature algorithm used to sign the certificate,example=SHA256withRSA"`
	PublicKeyAlgorithm string    `json:"public_key_algorithm,omitempty" jsonschema:"description=public key algorithm used in the certificate,example=RSA,example=ECDSA"`
	// Signature          string    `json:"signature,omitempty" jsonschema:"description=base64 encoded signature of the certificate"`
	SHA1Fingerprint   string `json:"sha1_fingerprint,omitempty" jsonschema:"description=SHA1 fingerprint of the certificate,example=3A:5B:7C:8D:9E:0F:1A:2B:3C:4D:5E:6F:7A:8B:9C"`
	SHA256Fingerprint string `json:"sha256_fingerprint,omitempty" jsonschema:"description=SHA256 fingerprint of the certificate,example=3A:5B:7C:8D:9E:0F:1A:2B:3C:4D"`
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

	Port                 uint16        `bun:"port,unique:port_protocol_addr_network_interface_id" json:"port" jsonschema:"description=port number,example=22,example=80,example=443,minimum=1,maximum=65535"`
	Protocol             string        `bun:"protocol,unique:port_protocol_addr_network_interface_id" json:"protocol" jsonschema:"description=transport layer protocol,example=tcp,example=udp"`
	Addr                 net.IP        `bun:"addr,unique:port_protocol_addr_network_interface_id" json:"addr" jsonschema:"description=IP address the application listens on,example=127.0.0.1,example=fe80::12b8:43e7:11f0:a406"`
	TLS                  *TLS          `bun:"tls,type:json" json:"tls,omitempty" jsonschema:"description=TLS information if the endpoint is using TLS"`
	Fingerprints         *Fingerprints `bun:"fingerprints,type:json" json:"fingerprints,omitempty" jsonschema:"description=application fingerprints"`
	ApplicationProtocols []string      `bun:"application_protocols,type:json" json:"application_protocols,omitempty" jsonschema:"description=list of application layer protocols detected on this endpoint,example=[\"http\",\"http/2\"]"`

	ApplicationID int64        `bun:"application_id,nullzero"`
	Application   *Application `bun:"rel:belongs-to,join:application_id=id"`

	NetworkInterfaceID int64             `bun:"network_interface_id,unique:port_protocol_addr_network_interface_id"` // can be null
	NetworkInterface   *NetworkInterface `bun:"rel:belongs-to,join:network_interface_id=id"`
}

func (s *Application) lastEndpoint() *ApplicationEndpoint {
	if len(s.Endpoints) == 0 {
		return nil
	}
	return s.Endpoints[len(s.Endpoints)-1]
}

// AddEndpoint appends a new endpoint if it does exist yet
// It returns true if a new endpoint has been added
func (s *Application) AddEndpoint(addr net.IP, port uint16, proto string) (*ApplicationEndpoint, bool) {
	// check if it exist
	for _, e := range s.Endpoints {
		if e.Port == port && e.Protocol == proto {
			// case where the application already listens to 0.0.0.0 (or ::)
			if e.Addr.IsUnspecified() {
				return e, false
			}
			// case where the incoming endpoint is unspecified
			// so it must encompass the first
			if addr.IsUnspecified() {
				copy(e.Addr, addr)
				return e, false
			}
		}
	}

	s.Endpoints = append(s.Endpoints,
		&ApplicationEndpoint{Addr: addr, Port: port, Protocol: proto})

	return s.lastEndpoint(), true
}

// func (s *Application) AddFlow(flow *Flow) {
// 	// check if it exist
// 	for _, f := range s.Flows {
// 		if f.Equal(flow) {
// 			return
// 		}
// 	}
// 	s.Flows = append(s.Flows, flow)
// }

func (ApplicationEndpoint) JSONSchemaExtend(schema *jsonschema.Schema) {
	if addrSchema, ok := schema.Properties.Get("addr"); ok {
		addrSchema.AnyOf = []*jsonschema.Schema{
			{Type: "string", Format: "ipv4"},
			{Type: "string", Format: "ipv6"},
		}
	}
}

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

// func (Flow) JSONSchema() *jsonschema.Schema {
// 	properties := orderedmap.New[string, *jsonschema.Schema]()

// 	for _, prop := range []string{"local_port", "remote_port"} {
// 		properties.Set(prop, &jsonschema.Schema{
// 			Type:        "integer",
// 			Maximum:     "65535",
// 			Minimum:     "1",
// 			Description: "port",
// 			Examples: []interface{}{
// 				22,
// 				80,
// 				443,
// 				49667,
// 			},
// 		})
// 	}

// 	for _, prop := range []string{"local_addr", "remote_addr"} {
// 		properties.Set(prop, &jsonschema.Schema{
// 			Title: "IPv4 or IPv6 address",
// 			AnyOf: []*jsonschema.Schema{
// 				{Type: "string", Format: "ipv4"},
// 				{Type: "string", Format: "ipv6"},
// 			},
// 			Description: "binding IP address",
// 			Examples: []interface{}{
// 				"192.168.10.103",
// 				"0.0.0.0",
// 				"::",
// 				"fe80::c1b2:a320:f799:10e0",
// 			},
// 		})
// 	}

// 	properties.Set("protocol", &jsonschema.Schema{
// 		Type:        "string",
// 		Description: "transport layer protocol",
// 		Examples: []interface{}{
// 			"tcp",
// 			"udp",
// 		},
// 	})

// 	properties.Set("status", &jsonschema.Schema{
// 		Type:        "string",
// 		Description: "Link status",
// 		Examples: []interface{}{
// 			"CLOSE_WAIT",
// 			"ESTABLISHED",
// 		},
// 	})

// 	properties.Set("remote_extra", jsonschema.Reflect(&FlowRemoteExtra{}))

// 	return &jsonschema.Schema{
// 		Properties:           properties,
// 		AdditionalProperties: jsonschema.FalseSchema,
// 		Type:                 "object",
// 		Required:             []string{"local_port", "local_addr", "remote_port", "remote_addr", "protocol", "status"},
// 	}
// }
