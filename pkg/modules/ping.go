// LINUX(PingModule) ok
// WINDOWS(PingModule) ok
// MACOS(PingModule) ?
// ROOT(PingModule) no
package modules

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/asiffer/puzzle"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/modules/ping"
	"github.com/situation-sh/situation/pkg/store"
	"github.com/situation-sh/situation/pkg/utils"
)

func init() {
	registerModule(&PingModule{
		timeout: 300 * time.Millisecond,
	})
}

// PingModule pings local networks to discover new hosts.
//
// The module relies on [pro-bing] library.
//
// A single ping attempt is made on every host of the local networks
// (the host may belong to several networks). Only IPv4 networks with
// prefix length >=20 are treated.
// The ping timeout is hardset to 300ms.
//
// [pro-bing]: https://github.com/prometheus-community/pro-bing
type PingModule struct {
	BaseModule

	timeout time.Duration
}

func (m *PingModule) Bind(config *puzzle.Config) error {
	return setDefault(config, m, "timeout", &m.timeout, "Ping timeout")
}

func (m *PingModule) Name() string {
	return "ping"
}

func (m *PingModule) Dependencies() []string {
	return []string{"host-network"}
}

func pingSubnetwork(ctx context.Context, network *net.IPNet, subnetID int64, source net.IP, logger logrus.FieldLogger, s *store.BunStorage) error {
	ips := utils.ListIPs(network)

	ipChan := make(chan net.IP, len(ips))

	onRecv := func(addr net.IP) {
		ipChan <- addr
		logger.WithField("ip", addr).Infof("Host found")
	}
	pool := utils.NewIPWorkerPool(64, func(ip net.IP) error {
		// return singlePing(ip, onRecv, logger)
		return ping.Ping4(ip, source, 500*time.Millisecond, onRecv)
	})

	if err := pool.Run(ips); err != nil {
		logger.
			WithField("network", network).
			WithError(err).
			Warnf("error while pinging subnetwork")
	}
	close(ipChan)

	subnet := s.GetOrCreateSubnetwork(ctx, network.String())
	if subnet == nil {
		return fmt.Errorf("unable to create or retrieve subnetwork %s", network.String())
	}
	nics := make([]models.NetworkInterface, 0)
	for ip := range ipChan {
		nics = append(nics, models.NetworkInterface{
			IP: ip.String(),
		})
	}
	_, err := s.DB().
		NewInsert().
		Model(&nics).
		On("CONFLICT (ip, subnetwork_id) DO UPDATE").
		Set("updated_at = CURRENT_TIMESTAMP").
		Exec(ctx)

	links := make([]models.NetworkInterfaceSubnet, 0)
	for _, nic := range nics {
		link := models.NetworkInterfaceSubnet{
			NetworkInterfaceID: nic.ID,
			SubnetworkID:       subnet.ID,
		}
		links = append(links, link)
	}
	_, err = s.DB().
		NewInsert().
		Model(&links).
		On("CONFLICT DO NOTHING").
		Exec(ctx)

	if err != nil {
		// fmt.Println("nics:", nics)
		return fmt.Errorf("unable to insert new NICs for subnetwork %s: %v", network.String(), err)
	}
	return nil
}

// Ping sends unprivileged ICMP echo messages to all
// hosts on a subnetwork
func (m *PingModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	// host := store.GetHost()
	// try to ping all networks
	for _, network := range storage.GetAllIPv4Networks(ctx) {
		// network() returns the IPv4 network attached to this nic
		// for _, network := range []*net.IPNet{nic.Network()} {
		// if network == nil {
		// 	continue
		// }
		ipnet, err := network.IPNet()
		_, zeros := ipnet.Mask.Size()

		if err != nil {
			logger.
				WithField("network", network.NetworkCIDR).
				WithError(err).
				Warn("unable to parse network CIDR")
			continue
		}

		switch ones := network.MaskSize; {
		case ones < 20:
			// ignore to large network (here /20 at most)
			logger.Warnf("Ignoring %s (network is too wide)", ipnet)
			continue
		case ones > 24:
			// if the network is restricted. We try to
			// send pings in a wider one. It may appear
			// in VPN cases (so we ensure that the base ip is not public)
			// this change does not modify the mask inside the store
			if !utils.IsPublic(ipnet.IP) {
				ipnet.Mask = net.CIDRMask(24, zeros)
			}

		}

		logger.Infof("Pinging %s", ipnet)
		if err := pingSubnetwork(ctx, ipnet, network.ID, nil, logger, storage); err != nil {
			logger.
				WithField("network", network).
				WithError(err).
				Error("error while pinging subnetwork")
		}
		// }
	}

	return nil
}
