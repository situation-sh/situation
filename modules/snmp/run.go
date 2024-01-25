package snmp

import (
	"net"
	"strings"
	"sync"

	"github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/store"
)

func RunSingle(g *gosnmp.GoSNMP, m *models.Machine, wg *sync.WaitGroup, cerr chan error, logger *logrus.Entry) {
	defer wg.Done()
	if err := g.Connect(); err != nil {
		cerr <- err
		return
	}
	defer g.Conn.Close()

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
	remote := store.GetMachineByIP(remoteIP)
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
		if iface.gateway() == nil {
			continue
		}

		for _, nic := range m.NICS { // loop over all the machine's nic
			// check MAC or Name match
			if nic.Match(nil, iface.MAC) || nic.Name == iface.Name {
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

			if nic.IP != nil {
				logger = logger.WithField("ip", nic.IP).WithField("mask_size", nic.MaskSize)
			}

			if nic.IP6 != nil {
				logger = logger.WithField("ip6", nic.IP6).WithField("mask6_size", nic.Mask6Size)
			}
			logger.Info("Network Interface found on network")
		}
	}
	// add interfaces
	m.NICS = append(m.NICS, ifaceToAdd...)
}
