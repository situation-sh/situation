// LINUX(ARPModule) ok
// WINDOWS(ARPModule) ok
// MACOS(ARPModule) ?
// ROOT(ARPModule) no
package modules

import (
	"fmt"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/modules/arp"
)

func init() {
	registerModule(&ARPModule{})
}

// ARPModule reads internal ARP table to find network neighbors.
// It **does not send ARP requests** but leverage the [Ping] module
// that is likely to update the local table.
//
// On Linux, it uses the Netlink API with the [netlink] library.
// On Windows, it calls `GetIpNetTable2`.
//
// [Ping]: ping.md
//
// [netlink]: https://github.com/vishvananda/netlink1
type ARPModule struct {
	BaseModule
}

func (m *ARPModule) Name() string {
	return "arp"
}

func (m *ARPModule) Dependencies() []string {
	// depends on ping to ensure a rather fresh
	// arp table
	return []string{"ping"}
}

func (m *ARPModule) Run() error {

	m.logger.Info("Retrieving ARP table")
	table, err := arp.GetARPTable()
	if err != nil {
		return err
	}

	host := m.store.GetHost()
	if host == nil {
		return fmt.Errorf("no host found")
	}

	for _, nic := range host.NICS {

		if nic.Network() == nil {
			continue
		}

		// fmt.Printf("%+v\n", nic.Network())
		// fmt.Printf("%+v\n", arp.FilterARPTableByNetwork(table, nic.Network()))

		// ipv4 --------------------------------------------------------------
		for _, entry := range arp.FilterARPTableByNetwork(table, nic.Network()) {
			// normally the arp module come after the ping module
			// so we are likely to already get a machine with the
			// entry IP
			machine := m.store.GetMachineByIP(entry.IP)

			if machine != nil {
				// check mac address
				nic := machine.GetNetworkInterfaceByIP(entry.IP)
				// fmt.Printf("%+v [%v]\n", nic, nic.MAC == nil)
				if len(nic.MAC) == 0 {
					// assign the new mac
					nic.MAC = entry.MAC
					m.logger.WithField("ip", nic.IP).WithField("mac", nic.MAC).Info("MAC address assigned")
				} else if nic.MAC.String() != entry.MAC.String() {
					return fmt.Errorf("MAC address mismatch for machine with ip %v (%v != %v)",
						entry.IP, nic.MAC, entry.MAC)
				}
				continue
			}

			machine = m.store.GetMachineByMAC(entry.MAC)
			if machine != nil {
				// check ip address
				nic := machine.GetNetworkInterfaceByMAC(entry.MAC)
				if nic.IP == nil {
					// assign the new IP
					nic.IP = entry.IP.To4()
					m.logger.WithField("ip", nic.IP).WithField("mac", nic.MAC).Debug("IP address assigned")
				} else if !nic.IP.Equal(entry.IP) {
					return fmt.Errorf("IP address mismatch for machine with mac %v (%v != %v)",
						entry.MAC, nic.IP, entry.IP)
				}
				continue
			}

			// if you reach this code, it means that there is no machine
			// with this IP or this MAC. So we can create it!
			machine = models.NewMachine()
			machine.NICS = append(machine.NICS, &models.NetworkInterface{
				MAC:      entry.MAC,
				IP:       entry.IP.To4(),
				MaskSize: nic.MaskSize,
			})

			m.logger.WithField("mac", machine.NICS[0].MAC).
				WithField("ip", machine.NICS[0].IP).
				WithField("mask", machine.NICS[0].MaskSize).
				Info("New machine added")
			// put this machine to the store
			m.store.InsertMachine(machine)

		}

	}

	return nil
}
