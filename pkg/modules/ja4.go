// LINUX(JA4Module) ok
// WINDOWS(JA4Module) ok
// MACOS(JA4Module) ?
// ROOT(JA4Module) no
package modules

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/modules/ja4"
)

func init() {
	registerModule(&JA4Module{})
}

// Module definition ---------------------------------------------------------

// JA4Module attempts JA4, JA4S and JA4X fingerprinting
//
// For technical details you look at https://github.com/FoxIO-LLC/ja4/blob/main/technical_details/README.md
// It first look at TLS endpoints (given by the [TLS module](./tls.md)) and then tries to connect to them,
// collecting then JA4, JA4S and JA4X fingerprints.
type JA4Module struct {
	BaseModule
}

func (m *JA4Module) Name() string {
	return "ja4"
}

func (m *JA4Module) Dependencies() []string {
	return []string{"tls"}
}

func (m *JA4Module) Run() error {

	for machine := range m.store.IterateMachines() {
		for _, app := range machine.Applications() {
			for _, endpoint := range app.Endpoints {
				if endpoint.TLS == nil {
					continue
				}
				if ja4, err := m.fingerprint(endpoint.Protocol, endpoint.Addr, endpoint.Port, m.logger); err == nil && ja4 != nil {
					if endpoint.Fingerprints == nil {
						endpoint.Fingerprints = &models.Fingerprints{}
					}
					endpoint.Fingerprints.JA4 = ja4
				}
			}

			for _, flow := range app.Flows {
				if flow.RemoteExtra == nil {
					continue
				}
				if flow.RemoteExtra.TLS == nil {
					continue
				}
				if ja4, err := m.fingerprint(flow.Protocol, flow.RemoteAddr, flow.RemotePort, m.logger); err == nil && ja4 != nil {
					if flow.RemoteExtra.Fingerprints == nil {
						flow.RemoteExtra.Fingerprints = &models.Fingerprints{}
					}
					flow.RemoteExtra.Fingerprints.JA4 = ja4
				}
			}
		}
	}
	return nil
}

func (m *JA4Module) fingerprint(proto string, ip net.IP, port uint16, logger logrus.FieldLogger) (*models.JA4, error) {
	target := net.JoinHostPort(ip.String(), fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout(
		"tcp",
		target,
		1*time.Second,
	)
	if err != nil {
		m.logger.WithError(err).
			WithField("ip", ip).
			WithField("port", port).
			WithField("protocol", proto).
			Error("fail to dial")
		return nil, err
	}

	tconn := &tapConn{Conn: conn}
	tlsConn := tls.Client(tconn, &tls.Config{
		InsecureSkipVerify: true, // #nosec G402 - skip certificate verification for scanning
	})
	err = tlsConn.Handshake()
	if err != nil {
		m.logger.WithError(err).
			WithField("ip", ip).
			WithField("port", port).
			WithField("protocol", proto).
			Error("TLS handshake failed")
		return nil, err
	}

	ja4_, err := ja4.JA4(tconn.writeBuf)
	if err != nil {
		m.logger.WithError(err).
			WithField("ip", ip).
			WithField("port", port).
			WithField("protocol", proto).
			Error("JA4 parse failed")
		return nil, err
	}

	// Step 3: Extract JA4S from raw handshake
	ja4s, err := ja4.JA4S(tconn.readBuf)
	if err != nil {
		m.logger.WithError(err).
			WithField("ip", ip).
			WithField("port", port).
			WithField("protocol", proto).
			Error("JA4S parse failed")
		return nil, err
	}

	fp := models.JA4{
		JA4:  ja4_,
		JA4S: ja4s,
	}

	entry := m.logger.
		WithField("ip", ip).
		WithField("port", port).
		WithField("protocol", proto).
		WithField("ja4", fp.JA4).
		WithField("ja4s", fp.JA4S)

	// ja4x
	certs := tlsConn.ConnectionState().PeerCertificates
	if len(certs) > 0 {
		fp.JA4X = ja4.JA4X(certs[0])
		entry = entry.WithField("ja4x", fp.JA4X)
	}
	entry.Info("JA4 fingerprints retrieved")

	return &fp, nil

}

// tapConn saves handshake bytes
type tapConn struct {
	net.Conn
	readBuf  []byte
	writeBuf []byte
}

func (t *tapConn) Read(p []byte) (int, error) {
	n, err := t.Conn.Read(p)
	if n > 0 {
		t.readBuf = append(t.readBuf, p[:n]...)
	}
	return n, err
}

func (t *tapConn) Write(p []byte) (int, error) {
	n, err := t.Conn.Write(p)
	if n > 0 {
		t.writeBuf = append(t.writeBuf, p[:n]...)
	}
	return n, err
}
