// LINUX(ARPModule) ok
// WINDOWS(ARPModule) ok
// MACOS(ARPModule) ?
// ROOT(ARPModule) no
package modules

import (
	"context"
	"fmt"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/modules/arp"
	"github.com/situation-sh/situation/pkg/utils"
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
	// return []string{"host-network"}
}

func (m *ARPModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	hostID := storage.GetHostID(ctx)

	logger.Info("Retrieving ARP table")
	table, err := arp.GetARPTable()
	if err != nil {
		// logger.WithError(err).Error("Cannot retrieve ARP table")
		return fmt.Errorf("cannot retrieve ARP table: %w", err)
	}

	newNICS := make([]*models.NetworkInterface, 0)
	nicSubnetMapper := make(map[string]int64) // key: mac+subnetID, value: nicID
	// fmt.Println(storage.GetMachineNICs(ctx, hostID))

	// for every nic, we first find the neighbors (or we create them)
	// then we bind all these neighbors to the same subnetwork if it is not
	// the case yetS
	for _, nic := range storage.GetMachineNICs(ctx, hostID) {
		for _, network := range nic.Subnetworks {
			if network == nil {
				continue
			}

			logger.
				WithField("name", nic.Name).
				WithField("mac", nic.MAC).
				WithField("ip", nic.IP).
				WithField("network", network.NetworkCIDR).
				Debug("Looking for neighbors")
			// network := nic.Network()

			// orphanNICs := make([]*models.NetworkInterface, 0)

			// subnetID := nic.SubnetworkID
			// if nic.SubnetworkID == 0 {
			// 	orphanNICs = append(orphanNICs, &nic)
			// 	logger.
			// 		WithField("network", network).
			// 		Debug("No subnetwork exists yet")
			// }

			ipnet, err := network.IPNet()
			if err != nil {
				logger.
					WithError(err).
					WithField("network", network.NetworkCIDR).
					Warn("Cannot parse network CIDR")
				continue
			}

			for _, entry := range arp.FilterARPTableByNetwork(table, ipnet) {
				logger.
					WithField("mac", entry.MAC).
					WithField("ip", entry.IP).
					Debug("Processing ARP entry")
				// we try to find the corresponding NIC object
				// var obj *models.NetworkInterface
				mac := entry.MAC.String()
				ip := entry.IP.String()
				obj := storage.GetNICByMACOrIPOnSubnet(ctx, mac, ip, network.ID)

				if obj != nil {
					if obj.MAC != mac {
						obj.MAC = mac
					}
					if !utils.Includes(obj.IP, ip) {
						obj.IP = append(obj.IP, ip)
					}
					newNICS = append(newNICS, obj)
				} else {
					// here we do not have the nic yet
					nic := models.NetworkInterface{
						MAC:   mac,
						IP:    []string{entry.IP.String()},
						Flags: models.NetworkInterfaceFlags{Up: true},
					}

					newNICS = append(newNICS, &nic)
					key := fmt.Sprintf("%v,%v", mac, ip)
					// fmt.Println("key0:", key)
					nicSubnetMapper[key] = network.ID

					logger.WithField("mac", entry.MAC).
						WithField("ip", entry.IP).
						Info("New machine added from ARP entry")
				}

			}

		}
	}

	if len(newNICS) > 0 {
		logger.
			WithField("nics", len(newNICS)).
			Info("New NICs found from ARP table")
		// insert new NICs
		err = storage.DB().
			NewInsert().
			Model(&newNICS). // id are scanned automatically (https://bun.uptrace.dev/guide/query-insert.html#bulk-insert)
			On("CONFLICT DO UPDATE").
			Set("updated_at = CURRENT_TIMESTAMP").
			Set("mac = EXCLUDED.mac").
			Set("ip = EXCLUDED.ip").
			Scan(ctx)
		if err != nil {
			logger.
				WithError(err).
				WithField("nics", len(newNICS)).
				Error("Cannot insert new NICs")
			return err
		}

		// fmt.Println("nicSubnetMapper:", nicSubnetMapper)
		// create links between NICs and subnetworks
		links := make([]models.NetworkInterfaceSubnet, 0)
		for _, nic := range newNICS {
			// fmt.Println(nic.MAC, nic.IP)
			for _, ip := range nic.IP {
				key := fmt.Sprintf("%v,%v", nic.MAC, ip)
				// fmt.Println("key1:", key)
				if subnetID, ok := nicSubnetMapper[key]; ok {
					link := models.NetworkInterfaceSubnet{
						NetworkInterfaceID: nic.ID,
						SubnetworkID:       subnetID,
					}
					links = append(links, link)
				}
			}
			// for _, network := range nic.Subnetworks {
			// 	// key := fmt.Sprintf("%v,%v", network.NetworkCIDR, nic.MAC)
			// 	key := fmt.Sprintf("%v,%v", nic.MAC, network.ID)
			// 	fmt.Println("key:", key)
			// 	if _, ok := nicSubnetMapper[key]; ok {
			// 		link := models.NetworkInterfaceSubnet{
			// 			NetworkInterfaceID: nic.ID,
			// 			SubnetworkID:       network.ID,
			// 		}
			// 		links = append(links, link)
			// 	}
			// }
		}
		// fmt.Println("LINKS:", links)
		if len(links) == 0 {
			logger.Warn("No NIC - subnetwork links to insert")
			return nil
		}

		_, err = storage.DB().
			NewInsert().
			Model(&links).
			On("CONFLICT DO NOTHING").
			Exec(ctx)
		if err != nil {
			logger.WithError(err).
				WithField("links", len(links)).
				Error("Cannot insert new NIC - subnetwork links")
			return err
		}
		logger.
			WithField("nics", len(newNICS)).
			WithField("links", len(links)).
			Info("Inserted new NICs and NIC - subnetwork links")

	} else {
		logger.Info("No new NICs found from ARP table")
	}

	return nil
}
