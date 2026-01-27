// LINUX(FingerprintModule) ok
// WINDOWS(FingerprintModule) ok
// MACOS(FingerprintModule) ?
// ROOT(FingerprintModule) no
package modules

import (
	"context"
	"net"
	"os"
	"strings"

	"github.com/cakturk/go-netstat/netstat"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/store"
)

func init() {
	registerModule(&FingerprintModule{})
}

// FingerprintModule attempts to match the local host against machines
// already discovered in the shared database.
//
// This module is critical for multi-agent deployments where Agent A
// may have discovered Host B remotely (via ARP, ping, TCP scan), and
// later Agent B starts on Host B. Without fingerprinting, Agent B would
// create a duplicate machine entry instead of recognizing itself.
//
// Matching strategy:
//  1. Agent UUID match → definitive (reconnection case)
//  2. HostID (system UUID) match → definitive
//  3. Fuzzy matching on MAC/IP/hostname with weighted scores
//
// The module runs before any other module (no dependencies) to ensure
// the host machine is correctly identified before other modules populate it.
type FingerprintModule struct {
	BaseModule
}

func (m *FingerprintModule) Name() string {
	return "fingerprint"
}

func (m *FingerprintModule) Dependencies() []string {
	// No dependencies - this module must run first
	return nil
}

func (m *FingerprintModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)
	agent := getAgent(ctx)

	// Collect local fingerprint data
	query, err := collectFingerprintQuery(agent)
	if err != nil {
		logger.WithError(err).Warn("Failed to collect local fingerprint")
		return err
	}

	logger.
		WithField("host_id", query.HostID).
		WithField("macs", query.MACs).
		WithField("ips", query.IPs).
		WithField("hostname", query.Hostname).
		WithField("ports", len(query.Ports)).
		Debug("Local fingerprint collected")

	// Try to find a matching machine in the database
	match, err := storage.FindMachineByFingerprint(ctx, query)
	if err != nil {
		logger.WithError(err).Warn("Failed to search for fingerprint match")
		// Don't fail the module, just continue without matching
	}

	if match != nil {
		logger.
			WithField("machine_id", match.Machine.ID).
			WithField("score", match.Score).
			WithField("definitive", match.IsDefinitive).
			WithField("matched_on", match.MatchedOn).
			Info("Found matching machine in database, claiming it")

		// Claim this machine by setting the agent (if not already set)
		if match.Machine.Agent != agent {
			_, err := storage.DB().
				NewUpdate().
				Model((*models.Machine)(nil)).
				Where("id = ?", match.Machine.ID).
				Set("agent = ?", agent).
				Set("updated_at = CURRENT_TIMESTAMP").
				Exec(ctx)

			if err != nil {
				logger.WithError(err).Error("Failed to claim machine")
				return err
			}
		}

		// Update cache so other modules use the correct host ID
		storage.SetHostID(match.Machine.ID)

		logger.
			WithField("machine_id", match.Machine.ID).
			WithField("agent", agent).
			WithField("score", match.Score).
			Info("Successfully claimed existing machine")
	} else {
		logger.Info("No matching machine found, a new host will be created")
	}

	return nil
}

// collectFingerprintQuery gathers identifiers from the local machine
func collectFingerprintQuery(agent string) (*store.FingerprintQuery, error) {
	query := &store.FingerprintQuery{
		Agent: agent,
		MACs:  make([]string, 0),
		IPs:   make([]string, 0),
		Ports: make([]uint16, 0),
	}

	// Get hostname
	if h, err := os.Hostname(); err == nil {
		query.Hostname = h
	}

	// Get HostID (system UUID)
	if info, err := host.Info(); err == nil {
		query.HostID = info.HostID
	}

	// Get network interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		// Skip interfaces that are down
		if (iface.Flags & net.FlagUp) == 0 {
			continue
		}

		// Skip loopback
		if (iface.Flags & net.FlagLoopback) != 0 {
			continue
		}

		// Skip virtual interfaces (veth, virbr, etc.)
		if strings.HasPrefix(iface.Name, "veth") ||
			strings.HasPrefix(iface.Name, "virbr") ||
			strings.Contains(iface.Name, "qemu") {
			continue
		}

		// Collect MAC address
		if len(iface.HardwareAddr) > 0 {
			mac := iface.HardwareAddr.String()
			if mac != "" {
				query.MACs = append(query.MACs, mac)
			}
		}

		// Collect IP addresses
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			// Skip loopback and link-local addresses
			if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}

			query.IPs = append(query.IPs, ip.String())
		}
	}

	// Get listening ports (TCP and UDP)
	query.Ports = collectListeningPorts()

	return query, nil
}

// collectListeningPorts returns a list of TCP/UDP ports the machine is listening on
func collectListeningPorts() []uint16 {
	ports := make(map[uint16]bool)

	// Filter for listening sockets only
	listenFilter := func(e *netstat.SockTabEntry) bool {
		return e.State == netstat.Listen
	}

	// Collect TCP listening ports
	if entries, err := netstat.TCPSocks(listenFilter); err == nil {
		for _, entry := range entries {
			if entry.LocalAddr.Port > 0 {
				ports[entry.LocalAddr.Port] = true
			}
		}
	}

	// Collect TCP6 listening ports
	if entries, err := netstat.TCP6Socks(listenFilter); err == nil {
		for _, entry := range entries {
			if entry.LocalAddr.Port > 0 {
				ports[entry.LocalAddr.Port] = true
			}
		}
	}

	// IGNORE UDP for the moment
	// UDP doesn't have "listen" state, but we can get bound ports
	// by accepting all entries (UDP sockets are always "unconnected" when listening)
	// udpFilter := func(e *netstat.SockTabEntry) bool {
	// 	return true
	// }

	// if entries, err := netstat.UDPSocks(udpFilter); err == nil {
	// 	for _, entry := range entries {
	// 		if entry.LocalAddr.Port > 0 {
	// 			ports[entry.LocalAddr.Port] = true
	// 		}
	// 	}
	// }

	// if entries, err := netstat.UDP6Socks(udpFilter); err == nil {
	// 	for _, entry := range entries {
	// 		if entry.LocalAddr.Port > 0 {
	// 			ports[entry.LocalAddr.Port] = true
	// 		}
	// 	}
	// }

	// Convert map to slice
	result := make([]uint16, 0, len(ports))
	for port := range ports {
		result = append(result, port)
	}

	return result
}
