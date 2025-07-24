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

	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/modules/ja4"
	"github.com/situation-sh/situation/store"
)

func init() {
	RegisterModule(&JA4Module{})
}

// Module definition ---------------------------------------------------------

// JA4Module attempts JA4, JA4S and JA4X fingerprinting
//
// For technical details you look at https://github.com/FoxIO-LLC/ja4/blob/main/technical_details/README.md
// It first look at TLS endpoints (given by the [TLS module](./tls.md)) and then tries to connect to them,
// collecting then JA4, JA4S and JA4X fingerprints.
type JA4Module struct{}

func (m *JA4Module) Name() string {
	return "ja4"
}

func (m *JA4Module) Dependencies() []string {
	return []string{"tls"}
}

func (m *JA4Module) Run() error {
	logger := GetLogger(m)
	for machine := range store.IterateMachines() {
		for _, app := range machine.Applications() {
			for _, endpoint := range app.Endpoints {
				if endpoint.TLS == nil {
					continue
				}

				target := net.JoinHostPort(endpoint.Addr.String(), fmt.Sprintf("%d", endpoint.Port))
				conn, err := net.DialTimeout(
					"tcp",
					target,
					1*time.Second,
				)
				if err != nil {
					logger.WithError(err).
						WithField("ip", endpoint.Addr).
						WithField("port", endpoint.Port).
						WithField("protocol", endpoint.Protocol).
						Error("fail to dial")
					continue
				}

				tconn := &tapConn{Conn: conn}
				tlsConn := tls.Client(tconn, &tls.Config{
					InsecureSkipVerify: true, // #nosec G402 - skip certificate verification for scanning
				})
				err = tlsConn.Handshake()
				if err != nil {
					logger.WithError(err).
						WithField("ip", endpoint.Addr).
						WithField("port", endpoint.Port).
						WithField("protocol", endpoint.Protocol).
						Error("TLS handshake failed")
					continue
				}

				ja4_, err := ja4.JA4(tconn.writeBuf)
				if err != nil {
					logger.WithError(err).
						WithField("ip", endpoint.Addr).
						WithField("port", endpoint.Port).
						WithField("protocol", endpoint.Protocol).
						Error("JA4 parse failed")
					continue
				}

				// Step 3: Extract JA4S from raw handshake
				ja4s, err := ja4.JA4S(tconn.readBuf)
				if err != nil {
					logger.WithError(err).
						WithField("ip", endpoint.Addr).
						WithField("port", endpoint.Port).
						WithField("protocol", endpoint.Protocol).
						Error("JA4S parse failed")
					continue
				}

				fp := models.JA4{
					JA4:  ja4_,
					JA4S: ja4s,
				}

				entry := logger.
					WithField("ip", endpoint.Addr).
					WithField("port", endpoint.Port).
					WithField("protocol", endpoint.Protocol).
					WithField("ja4s", fp.JA4S)

				// ja4x
				certs := tlsConn.ConnectionState().PeerCertificates
				if len(certs) > 0 {
					fp.JA4X = ja4.JA4X(certs[0])
					entry = entry.WithField("ja4x", fp.JA4X)
				}

				if endpoint.Fingerprints == nil {
					endpoint.Fingerprints = &models.Fingerprints{}
				}

				endpoint.Fingerprints.JA4 = &fp
				entry.Info("JA4 fingerprints retrieved")
			}
		}
	}
	return nil
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
