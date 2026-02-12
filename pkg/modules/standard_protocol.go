// LINUX(StandardProtocolModule) ok
// WINDOWS(StandardProtocolModule) ok
// MACOS(StandardProtocolModule) ?
// ROOT(StandardProtocolModule) no
package modules

import (
	"context"
	"fmt"
	"strings"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/store"
	"github.com/uptrace/bun"
)

func init() {
	registerModule(&StandardProtocolModule{})
}

// StandardProtocolModule fills standard protocol information for endpoints.
type StandardProtocolModule struct {
	BaseModule
}

func (m *StandardProtocolModule) Name() string {
	return "standard-protocol"
}

func (m *StandardProtocolModule) Dependencies() []string {
	return []string{"netstat"}
}

func (m *StandardProtocolModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	// update TCP endpoints
	res, err := storage.DB().
		NewUpdate().
		Model((*models.ApplicationEndpoint)(nil)).
		Where("protocol = ?", "tcp").
		Where("application_protocols IS NULL").
		Where("port IN (?)", bun.In(stdPorts(stdTCPProtocols))).
		SetColumn("application_protocols", sqlCase(storage, stdTCPProtocols)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update standard tcp protocols: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get number of updated rows: %w", err)
	}
	logger.WithField("endpoints", n).Info("tcp endpoints updated")

	// update UDP endpoints
	res, err = storage.DB().
		NewUpdate().
		Model((*models.ApplicationEndpoint)(nil)).
		Where("protocol = ?", "udp").
		Where("application_protocols IS NULL").
		Where("port IN (?)", bun.In(stdPorts(stdUDPProtocols))).
		SetColumn("application_protocols", sqlCase(storage, stdUDPProtocols)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update standard udp protocols: %w", err)
	}
	n, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get number of updated rows: %w", err)
	}
	logger.WithField("endpoints", n).Info("udp endpoints updated")

	return nil
}

func sqlCase(storage *store.BunStorage, protocols map[uint16]string) string {
	chunks := []string{"( CASE port"}
	for port, proto := range protocols {
		chunks = append(chunks,
			fmt.Sprintf(`WHEN %d THEN '%s'`,
				port, storage.ARRAY([]string{proto}),
			),
		)
	}
	chunks = append(chunks, "END )")
	return strings.Join(chunks, "\n ")
}

func stdPorts(protocols map[uint16]string) []uint16 {
	ports := make([]uint16, 0, len(protocols))
	for port := range protocols {
		ports = append(ports, port)
	}
	return ports
}

var stdTCPProtocols = map[uint16]string{
	7:     "echo",
	9:     "discard",
	20:    "ftp-data",
	21:    "ftp",
	22:    "ssh",
	25:    "smtp",
	37:    "time",
	43:    "whois",
	53:    "dns",
	67:    "dhcp",
	68:    "dhcp",
	80:    "http",
	88:    "kerberos",
	110:   "pop3",
	111:   "onc-rpc",
	115:   "sftp",
	123:   "ntp",
	137:   "netbios-ns",
	139:   "netbios-ssn",
	143:   "imap",
	162:   "snmp",
	170:   "print-srv",
	179:   "bgp",
	194:   "irc",
	220:   "imap3",
	389:   "ldap",
	443:   "https",
	445:   "smb",
	465:   "smtp-tls",
	502:   "modbus",
	513:   "rlogin",
	515:   "printer",
	530:   "rpc",
	587:   "smtp-tls",
	631:   "ipp",
	636:   "ldap-tls",
	749:   "kerberos",
	853:   "dns-tls",
	989:   "ftps-data",
	990:   "ftps",
	992:   "telnet-tls",
	993:   "imap-tls",
	995:   "pop3-tls",
	1194:  "openvpn",
	1293:  "ipsec",
	1812:  "radius",
	1813:  "radius",
	1883:  "mqtt",
	2049:  "nfs",
	2083:  "radsec",
	2375:  "docker",
	2376:  "docker-tls",
	2377:  "docker-swarm",
	2775:  "smpp",
	3260:  "iscsi",
	3306:  "mysql",
	3389:  "rdp",
	3659:  "apple-sasl",
	5432:  "postgresql",
	5060:  "sip",
	5061:  "sip-tls",
	5173:  "vite",
	5222:  "xmpp-client",
	5355:  "llmnr",
	5357:  "wsdapi",
	5601:  "kibana",
	5670:  "zeromq",
	5671:  "amqp-tls",
	5672:  "amqp",
	5900:  "vnc",
	5984:  "couchdb",
	6379:  "redis",
	6432:  "pgbouncer",
	6514:  "syslog-tls",
	6653:  "openflow",
	6665:  "irc",
	6666:  "irc",
	6667:  "irc",
	6668:  "irc",
	6669:  "irc",
	6697:  "irc-tls",
	7474:  "neo4j",
	7687:  "boltdb",
	8006:  "proxmox",
	8080:  "http-alt",
	8089:  "splunk",
	8093:  "gitlab",
	8125:  "statsd",
	8222:  "vmware-http",
	8333:  "vmware-https",
	8443:  "https-alt",
	8530:  "windows-update-http",
	8531:  "windows-update-https",
	8883:  "mqtt-tls",
	8983:  "solr",
	9006:  "tomcat",
	9100:  "raw-print",
	9200:  "elasticsearch",
	10050: "zabbix-agent",
	10051: "zabbix-trapper",
	10514: "rsyslog-tls",
	11211: "memcached",
	11434: "ollama",
	11920: "syncthing",
	27017: "mongodb",
	32400: "plex",
}

var stdUDPProtocols = map[uint16]string{
	7:     "echo",
	9:     "discard",
	37:    "time",
	53:    "dns",
	67:    "dhcp",
	68:    "dhcp",
	69:    "tftp",
	88:    "kerberos",
	111:   "onc-rpc",
	123:   "ntp",
	137:   "netbios-ns",
	138:   "netbios-dgm",
	161:   "snmp",
	162:   "snmp-trap",
	389:   "ldap",
	443:   "quic",
	500:   "isakmp",
	514:   "syslog",
	520:   "rip",
	546:   "dhcpv6-client",
	547:   "dhcpv6-server",
	623:   "ipmi",
	631:   "ipp",
	749:   "kerberos",
	853:   "dns-dtls",
	1194:  "openvpn",
	1812:  "radius",
	1813:  "radius-acct",
	1900:  "ssdp",
	2049:  "nfs",
	3478:  "stun",
	3702:  "ws-discovery",
	4500:  "ipsec-nat",
	5060:  "sip",
	5353:  "mdns",
	5355:  "llmnr",
	6343:  "sflow",
	8125:  "statsd",
	11211: "memcached",
}
