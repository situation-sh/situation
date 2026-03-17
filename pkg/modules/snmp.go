// LINUX(SNMPModule) ok
// WINDOWS(SNMPModule) ok
// MACOS(SNMPModule) ?
// ROOT(SNMPModule) no
package modules

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/asiffer/puzzle"
	"github.com/gosnmp/gosnmp"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/modules/snmp"
	"github.com/situation-sh/situation/pkg/store"

	"github.com/sirupsen/logrus"
)

func init() {
	registerModule(&SNMPModule{
		Version:   uint8(gosnmp.Version2c),
		Community: "public",
		Timeout:   3 * time.Second,
		Transport: "udp",
		Port:      161,
	})
}

// SNMPModule collects network interface data from neighbors via SNMP.
//
// This module requires access to the OID tree `.1.3.6.1.2.1`.
// In case of snmpd, the configuration (snmpd.conf) should include:
//
// ```
// view systemonly included .1.3.6.1.2.1
// ```
type SNMPModule struct {
	BaseModule
	Version   uint8
	Community string
	Timeout   time.Duration
	Transport string
	Port      uint16
}

func (m *SNMPModule) Bind(config *puzzle.Config) error {
	if err := setDefault(config, m, "version", &m.Version, "SNMP version to use"); err != nil {
		return err
	}
	if err := setDefault(config, m, "community", &m.Community, "SNMP community to query"); err != nil {
		return err
	}
	if err := setDefault(config, m, "timeout", &m.Timeout, "SNMP query timeout"); err != nil {
		return err
	}
	if err := setDefault(config, m, "transport", &m.Transport, "TCP or UDP transport protocol"); err != nil {
		return err
	}
	if err := setDefault(config, m, "port", &m.Port, "Port to connect"); err != nil {
		return err
	}
	return nil
}

func (m *SNMPModule) Name() string {
	return "snmp"
}

func (m *SNMPModule) Dependencies() []string {
	// depends on arp to ensure a rather fresh arp table
	return []string{"arp"}
}

func (m *SNMPModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	nics, err := storage.GetNeighorNICS(ctx)
	if err != nil {
		return fmt.Errorf("cannot retrieve neighbor NICs: %w", err)
	}

	if len(nics) == 0 {
		logger.Warn("No neighbor NICs found, skipping SNMP scan")
		return nil
	}

	logger.WithField("nics", len(nics)).Info("Starting SNMP scan")

	var (
		mu             sync.Mutex
		wg             sync.WaitGroup
		errs           []error
		newNICs        []*models.NetworkInterface
		updateNICs     []*models.NetworkInterface
		snmpEndpoints  []*models.ApplicationEndpoint
		updateMachines []*models.Machine
	)

	for _, nic := range nics {
		// skip orphan NICs (no machine to attach new interfaces to)
		if nic.MachineID == 0 {
			continue
		}
		for _, ip := range nic.IPs() {
			if ip.IsLoopback() || ip.IsMulticast() {
				continue
			}
			wg.Add(1)
			go func(targetNIC *models.NetworkInterface, targetIP net.IP) {
				defer wg.Done()
				g := gosnmp.GoSNMP{
					Target:    targetIP.String(),
					Version:   gosnmp.SnmpVersion(m.Version),
					Context:   ctx,
					Retries:   2,
					Transport: m.Transport,
					Port:      m.Port,
					Timeout:   m.Timeout,
					Community: m.Community,
				}
				result, err := runSingle(ctx, &g, targetNIC, logger, storage)
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					errs = append(errs, err)
					return
				}
				newNICs = append(newNICs, result.newNICs...)
				updateNICs = append(updateNICs, result.updateNICs...)
				if result.snmpEndpoint != nil {
					snmpEndpoints = append(snmpEndpoints, result.snmpEndpoint)
				}
				if len(result.updateMachine) > 0 {
					updateMachines = append(updateMachines, result.updateMachine...)
				}
			}(nic, ip)
		}
	}

	wg.Wait()

	// insert new NICs discovered via SNMP
	if len(newNICs) > 0 {
		logger.WithField("nics", len(newNICs)).Info("Inserting new NICs found via SNMP")
		if _, err := storage.DB().
			NewInsert().
			Model(&newNICs).
			On("CONFLICT DO NOTHING").
			Exec(ctx); err != nil {
			logger.WithError(err).Error("Cannot insert new NICs from SNMP")
			errs = append(errs, err)
		}
	}

	// update existing NICs enriched via SNMP (name, gateway, IPs)
	if len(updateNICs) > 0 {
		logger.WithField("nics", len(updateNICs)).Info("Updating existing NICs from SNMP data")
		if _, err := storage.DB().
			NewUpdate().
			Model(&updateNICs).
			Column("name", "gateway", "ip").
			Bulk().
			Exec(ctx); err != nil {
			logger.WithError(err).Error("Cannot update NICs from SNMP")
			errs = append(errs, err)
		}
	}

	// register SNMP service endpoints (UDP/161)
	if len(snmpEndpoints) > 0 {
		logger.WithField("endpoints", len(snmpEndpoints)).Info("Registering SNMP endpoints")
		if _, err := storage.DB().
			NewInsert().
			Model(&snmpEndpoints).
			On("CONFLICT DO NOTHING").
			Exec(ctx); err != nil {
			logger.WithError(err).Error("Cannot insert SNMP endpoints")
			errs = append(errs, err)
		}
	}

	// update machines with system information discovered via SNMP
	if len(updateMachines) > 0 {
		logger.WithField("machines", len(updateMachines)).Info("Updating machines with SNMP system information")
		if _, err := storage.DB().
			NewUpdate().
			Model(&updateMachines).
			Column("hostname", "platform", "distribution", "distribution_family", "arch", "uptime", "chassis").
			Bulk().
			Exec(ctx); err != nil {
			logger.WithError(err).Error("Cannot update machines with SNMP system information")
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// snmpResult holds network interfaces discovered from a single SNMP target.
type snmpResult struct {
	newNICs       []*models.NetworkInterface
	updateNICs    []*models.NetworkInterface
	snmpEndpoint  *models.ApplicationEndpoint // nil if SNMP was unreachable
	updateMachine []*models.Machine
}

func checkSNMP(g *gosnmp.GoSNMP) bool {
	if _, err := g.Get([]string{"0.0"}); err != nil {
		return false
	}
	return true
}

// runSingle performs an SNMP walk on g.Target, discovers network interfaces,
// compares them against the machine's existing NICs in the database, and
// returns new/updated interfaces plus the SNMP service endpoint.
func runSingle(ctx context.Context, g *gosnmp.GoSNMP, targetNIC *models.NetworkInterface, logger logrus.FieldLogger, s *store.BunStorage) (snmpResult, error) {
	empty := snmpResult{
		newNICs:       make([]*models.NetworkInterface, 0),
		updateNICs:    make([]*models.NetworkInterface, 0),
		updateMachine: make([]*models.Machine, 0),
	}

	if err := g.Connect(); err != nil {
		return empty, err
	}
	defer g.Conn.Close()

	// ensure SNMP is reachable on this host
	if !checkSNMP(g) {
		return empty, nil
	}

	ifaces, _ := snmp.GetAllInterfaces(g)
	if len(ifaces) == 0 {
		return empty, nil
	}

	// load this machine's existing NICs for matching
	existingNICs := s.GetMachineNICs(ctx, targetNIC.MachineID)

	result := snmpResult{
		newNICs:    make([]*models.NetworkInterface, 0),
		updateNICs: make([]*models.NetworkInterface, 0),
		// register the SNMP service endpoint (UDP/161)
		snmpEndpoint: &models.ApplicationEndpoint{
			Port:               g.Port,
			Protocol:           g.Transport,
			Addr:               g.Target,
			NetworkInterfaceID: targetNIC.ID,
		},
		updateMachine: make([]*models.Machine, 0),
	}

	for _, iface := range ifaces {
		mac := iface.MAC.String()
		// ignore empty or zeroed MAC (loopback typically)
		if mac == "" || strings.HasPrefix(mac, "00:00:00") {
			continue
		}
		// ignore interfaces without a gateway (internal-only interfaces)
		if iface.Gateway() == "" {
			continue
		}

		nic0 := iface.ToNetworkInterface()

		// find a matching existing NIC by MAC or name
		var match *models.NetworkInterface
		for _, nic := range existingNICs {
			if nic.MAC == mac || nic.Name == iface.Name {
				match = nic
				break
			}
		}

		if match != nil {
			// enrich the existing NIC with discovered data
			if match.Name == "" {
				match.Name = nic0.Name
			}
			if match.Gateway == "" {
				match.Gateway = nic0.Gateway
			}
			for _, ip := range nic0.IP {
				found := false
				for _, existing := range match.IP {
					if existing == ip {
						found = true
						break
					}
				}
				if !found {
					match.IP = append(match.IP, ip)
				}
			}
			result.updateNICs = append(result.updateNICs, match)
		} else {
			nic0.MachineID = targetNIC.MachineID
			result.newNICs = append(result.newNICs, nic0)
			logger.WithField("name", nic0.Name).
				WithField("mac", nic0.MAC).
				WithField("ip", nic0.IP).
				Info("Network interface found via SNMP")
		}
	}

	if targetNIC.Machine != nil {
		if update, err := snmp.PopulateSystem(g, targetNIC.Machine); err != nil {
			logger.
				WithError(err).
				Warn("Cannot populate system information from SNMP")
		} else if update {
			result.updateMachine = append(result.updateMachine, targetNIC.Machine)
		}
	}

	return result, nil
}
