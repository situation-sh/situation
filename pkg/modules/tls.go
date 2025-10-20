// LINUX(TLSModule) ok
// WINDOWS(TLSModule) ok
// MACOS(TLSModule) ?
// ROOT(TLSModule) no
package modules

import (
	"crypto/sha1" // #nosec G505 - just for cert fingerprint
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/utils"
)

// +--------+--------------------+---------------------------------------------+
// | Port   | Protocol           | Description                                 |
// +--------+--------------------+---------------------------------------------+
// | 443    | HTTPS              | HTTP over TLS (secure web)                  |
// | 465    | SMTPS              | SMTP over TLS (deprecated but still used)   |
// | 993    | IMAPS              | IMAP over TLS                               |
// | 995    | POP3S              | POP3 over TLS                               |
// | 636    | LDAPS              | LDAP over TLS                               |
// | 3269   | LDAPS (GC)         | Global Catalog over TLS (Microsoft AD)      |
// | 990    | FTPS (Implicit)    | FTP over TLS                                |
// | 989    | FTPS (Data)        | FTP data channel over TLS                   |
// | 853    | DNS over TLS       | Secure DNS over TLS                         |
// | 4433   | QUIC (TLS handshake)| QUIC using TLS 1.3 handshake               |
// | 5671   | AMQPS              | AMQP over TLS (e.g., RabbitMQ)              |
// | 8883   | MQTT over TLS      | MQTT protocol with TLS                      |
// | 6697   | IRC over TLS       | IRC (Internet Relay Chat) with TLS          |
// | 2484   | Oracle TCPS        | Oracle database over TLS                    |
// +--------+--------------------+---------------------------------------------+
func init() {
	m := &TLSModule{Ports: []uint16{443, 465, 993, 995, 636, 3269, 990, 989, 853, 4433, 5671, 8883, 6697, 2484}}
	registerModule(m)
	// we do not set defaults for the moment because the puzzle library is not ready to handle []uint16
	// SetDefault(m, "ports", &m.Ports, "TCP ports to scan for TLS connections")
}

// TLSModule enrich endpoints with TLS information.
//
// The module only uses the Go standard library. Currently it only supports TLS over TCP.
type TLSModule struct {
	BaseModule
	Ports []uint16
}

func (m *TLSModule) Name() string {
	return "tls"
}

func (m *TLSModule) Dependencies() []string {
	return []string{"tcp-scan", "netstat"}
}

func (m *TLSModule) Run() error {

	for machine := range m.store.IterateMachines() {
		for _, app := range machine.Applications() {
			for _, endpoint := range app.Endpoints {
				if info, err := m.tlsInfo(endpoint.Protocol, endpoint.Addr, endpoint.Port, m.logger); err == nil && info != nil {
					endpoint.TLS = info
				}
			}

			// flows
			for _, flow := range app.Flows {
				if info, err := m.tlsInfo(flow.Protocol, flow.RemoteAddr, flow.RemotePort, m.logger); err == nil && info != nil {
					if flow.RemoteExtra == nil {
						flow.RemoteExtra = &models.FlowRemoteExtra{}
					}
					flow.RemoteExtra.TLS = info
				}
			}
		}
	}
	return nil
}

func (m *TLSModule) tlsInfo(proto string, ip net.IP, port uint16, logger logrus.FieldLogger) (*models.TLS, error) {
	if proto != "tcp" {
		return nil, fmt.Errorf("unsupported protocol %s for TLS", proto)
	}
	if !utils.Includes(m.Ports, port) {
		return nil, fmt.Errorf("port %d not in configured TLS ports", port)
	}

	tlsInfo, err := getTLS(proto, ip, port)
	if err != nil {
		m.logger.WithError(err).
			WithField("ip", ip).
			Error("failed to get TLS info")
		return nil, err
	}

	m.logger.WithField("ip", ip).
		WithField("port", port).
		WithField("subject", tlsInfo.Subject).
		WithField("issuer", tlsInfo.Issuer).
		Info("TLS information retrieved")
	return tlsInfo, nil
}

func formatFingerprint(hash []byte) string {
	fingerprint := hex.EncodeToString(hash)
	formatted := ""
	for i := 0; i < len(fingerprint); i += 2 {
		formatted += fingerprint[i:i+2] + ":"
	}
	formatted = formatted[:len(formatted)-1] // remove trailing colon
	return formatted
}

func getTLS(network string, ip net.IP, port uint16) (*models.TLS, error) {
	addr := net.JoinHostPort(ip.String(), fmt.Sprintf("%d", port))
	conn, err := tls.Dial(network, addr, &tls.Config{
		InsecureSkipVerify: true, // #nosec G402 - skip certificate verification for scanning
	})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}

	cert := certs[0] // Leaf certificate

	// Compute SHA fingerprints
	sha1fp := sha1.Sum(cert.Raw) // #nosec G401 - we just store sha1 cert fingerprint
	sha256fp := sha256.Sum256(cert.Raw)

	infos := models.TLS{
		Subject:            cert.Subject.String(),
		Issuer:             cert.Issuer.String(),
		NotBefore:          cert.NotBefore,
		NotAfter:           cert.NotAfter,
		SerialNumber:       formatFingerprint(cert.SerialNumber.Bytes()),
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
		// Signature:          formatFingerprint(cert.Signature),
		SHA1Fingerprint:   formatFingerprint(sha1fp[:]),
		SHA256Fingerprint: formatFingerprint(sha256fp[:]),
	}

	return &infos, nil
}
