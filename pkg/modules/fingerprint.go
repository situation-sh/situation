// LINUX(FingerprintModule) ok
// WINDOWS(FingerprintModule) ok
// MACOS(FingerprintModule) ?
// ROOT(FingerprintModule) no
package modules

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/shirou/gopsutil/v4/host"
	"github.com/situation-sh/situation/pkg/models"
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

	machine := models.Machine{Agent: agent}
	// first try with host ID
	if info, err := host.Info(); err == nil {
		err = storage.DB().
			NewSelect().
			Model(&machine).
			Where("host_id = ?", info.HostID).
			Scan(ctx)
		if err == nil && machine.ID != 0 {
			storage.SetHostID(machine.ID)
			// update the machine with the given agent
			machine.Agent = agent
			_, err = storage.DB().
				NewUpdate().
				Model(&machine).
				Where("id = ?", machine.ID).
				Set("agent = ?", agent).
				Set("updated_at = CURRENT_TIMESTAMP").
				Exec(ctx)
			// here we prefer return an error if the update fails
			return err

		} else {
			logger.
				WithError(err).
				Debug("Fail to get machine by host ID")
		}
	} else {
		logger.
			WithError(err).
			Warn("Fail to retrieve host infos")
	}

	// then try with agent ID
	err := storage.DB().
		NewSelect().
		Model(&machine).
		Where("agent = ?", agent).
		Scan(ctx)
	if err == nil && machine.ID != 0 {
		logger.
			WithField("machine_id", machine.ID).
			Info("Found machine by agent ID, claiming it")

		// update cache so other modules use the correct host ID
		storage.SetHostID(machine.ID)
	} else {
		logger.
			WithError(err).
			Debug("Fail to get machine by agent ID")
	}

	// now try with nics
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			mac := strings.ToLower(iface.HardwareAddr.String())
			if mac == "" {
				// do not try to detect without MAC
				continue
			}
			addrs, err := iface.Addrs()
			if err != nil {
				// do not try to detect without addresses
				continue
			}
			ips := make([]string, 0)
			for _, addr := range addrs {
				if ip, _, err := net.ParseCIDR(addr.String()); err == nil {
					// skip loopback and link-local addresses
					if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
						continue
					}
					ips = append(ips, ip.String())
				}
			}
			if len(ips) == 0 {
				// do not try to detect without IP
				continue
			}

			nics := storage.GetNICByMACAndIPs(ctx, mac, ips)
			if len(nics) == 1 {
				// single candidate found (ouf!)
				nic := nics[0]
				if nic.Machine == nil || nic.Machine.ID == 0 {
					// create empty machine
					err = storage.DB().
						NewInsert().
						Model(&machine).
						Scan(ctx)
					if err != nil {
						// we prefer returning an error here
						return fmt.Errorf("fail to create host machine: %v", err)
					}
					logger.WithField("id", machine.ID).Info("Host machine created")

					storage.SetHostID(machine.ID)

					// link nic to machine
					nic.Machine = &machine
					nic.MachineID = machine.ID
					_, err = storage.DB().
						NewUpdate().
						Model(nic).
						Where("id = ?", nic.ID).
						Set("machine_id = ?", machine.ID).
						Set("updated_at = CURRENT_TIMESTAMP").
						Exec(ctx)
					if err != nil {
						return fmt.Errorf("fail to link nic to machine: %v", err)
					}
					logger.
						WithField("mac", nic.MAC).
						WithField("ips", nic.IP).
						WithField("machine_id", machine.ID).
						Info("Host machine linked to NIC")
					return nil

				}
			} else if len(nics) == 0 {
				// no candidate found
				// host-basic will create it later
				logger.
					WithField("mac", mac).
					WithField("ips", ips).
					Warn("no matching network interfaces found in database")
			}
		}
	}

	logger.Warn("No matching machine found")
	return nil
}
