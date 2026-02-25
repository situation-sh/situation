package snmp

import (
	"context"
	"net"
	"strings"
	"sync"

	"github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/store"
)

func checkSNMP(g *gosnmp.GoSNMP) bool {
	if _, err := g.Get([]string{"0.0"}); err != nil {
		return false
	}
	return true
}

func RunSingle(ctx context.Context, g *gosnmp.GoSNMP, m *models.Machine, wg *sync.WaitGroup, cerr chan error, logger logrus.FieldLogger, s *store.BunStorage) {
	defer wg.Done()
	if err := g.Connect(); err != nil {
		cerr <- err
		return
	}
	defer g.Conn.Close()

	// ensure we do have a SNMP connection
	if !checkSNMP(g) {
		return
	}

	// retrieve all the network interfaces collected by
	// snmp
	ifaces, err := getAllInterfaces(g)
	if len(ifaces) == 0 {
		cerr <- err
		return
	}

	// update remote machine by adding a remote port -------------------------
	// udp/161
	remoteIP := net.ParseIP(g.Target)
	remote := s.GetMachineByIP(ctx, remoteIP)
	if remote != nil {
		remote.GetOrCreateApplicationByEndpoint(
			g.Port,
			g.Transport,
			remoteIP,
		)
	}
	// -----------------------------------------------------------------------

	ifaceToAdd := make([]*models.NetworkInterface, 0)
	for _, iface := range ifaces { // loop over all the found ifaces
		match := false
		// ignore empty mac (localhost generally)
		mac := iface.MAC.String()
		if mac == "" || strings.HasPrefix(mac, "00:00:00") {
			continue
		}

		// ignore iface without gateway (it may remove other internal interfaces)
		if iface.gateway() == "" {
			continue
		}

		for _, nic := range m.NICS { // loop over all the machine's nic
			// check MAC or Name match
			if nic.MAC == iface.MAC.String() || nic.Name == iface.Name {
				match = true
				nic0 := iface.toNetworkInterface()
				// update
				nic.Merge(nic0)
				break
			}
		}
		if !match {
			nic := iface.toNetworkInterface()
			ifaceToAdd = append(ifaceToAdd, nic)
			logger = logger.WithField("name", nic.Name).
				WithField("mac", nic.MAC)

			if len(nic.IP) > 0 {
				logger = logger.WithField("ip", nic.IP)
			}

			// if nic.IP6 != nil {
			// 	logger = logger.WithField("ip6", nic.IP6).WithField("mask6_size", nic.Mask6Size)
			// }
			logger.Info("Network Interface found on network")
		}
	}
	// add interfaces
	m.NICS = append(m.NICS, ifaceToAdd...)
}
