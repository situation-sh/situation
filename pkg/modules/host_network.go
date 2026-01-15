// LINUX(HostNetworkModule) ok
// WINDOWS(HostNetworkModule) ok
// MACOS(HostNetworkModule) ?
// ROOT(HostNetworkModule) no
package modules

import (
	"context"
	"fmt"
	"net"
	"strings"

	netroute "github.com/libp2p/go-netroute"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/utils"
)

var (
	GOOGLE = net.IPv4(8, 8, 8, 8)
)

func init() {
	registerModule(&HostNetworkModule{})
}

// Module definition ---------------------------------------------------------

// HostNetworkModule retrieves basic newtork information about the host:
// interfaces along with their mac, ip and mask (IPv4 and IPv6)
//
// It uses the [go] standard library.
//
// On Linux, it uses the Netlink API.
// On Windows, it calls `GetAdaptersAddresses`.
//
// [go]: https://pkg.go.dev/net
type HostNetworkModule struct {
	BaseModule
}

func (m *HostNetworkModule) Name() string {
	return "host-network"
}

func (m *HostNetworkModule) Dependencies() []string {
	return []string{"host-basic"}
}

func (m *HostNetworkModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	hostID := storage.GetHostID(ctx)

	nics := make([]models.NetworkInterface, 0)
	subnets := make([]models.Subnetwork, 0)
	subnetNICMapper := make(map[string]bool)

	ifaces, err := getInterfaces()
	if err != nil {
		return fmt.Errorf("error while getting network interfaces: %v", err)
	}

	// create nics
	for _, iface := range ifaces {
		nic := models.NetworkInterface{MachineID: hostID}
		// name
		nic.Name = iface.Name
		// mac
		nic.MAC = iface.HardwareAddr.String()

		// flags (NEW!)
		nic.SetFlags(iface.Flags)

		// logging
		entry := logger.
			WithField("name", nic.Name).
			WithField("mac", nic.MAC)
		// nic gw
		_, gwIP, err := gatewayWithSrc(iface.HardwareAddr, nil)
		if err == nil {
			nic.Gateway = gwIP.String()
			entry = entry.WithField("gateway", nic.Gateway)
		}

		// gateway
		// if index == iface.Index {
		// 	nic.Gateway = gw.String()
		// 	entry = entry.WithField("gateway", nic.Gateway)
		// }

		entry.Info("Network Interface found on host")

		// ip(s)
		addrs, err := iface.Addrs()
		if err != nil {
			// ignore
			continue
		}

		for _, addr := range addrs {
			ip, ipnet, err := net.ParseCIDR(addr.String())
			// ignore
			if err != nil {
				continue
			}

			ms := utils.MaskSize(ipnet)
			s := models.Subnetwork{
				NetworkCIDR: ipnet.String(),
				NetworkAddr: ipnet.IP.String(),
				MaskSize:    ms,
			}

			// define subnet gateway if applicable
			_, gwIP, err := gatewayWithSrc(iface.HardwareAddr, ip)
			if err == nil && ipnet.Contains(gwIP) {
				s.Gateway = gwIP.String()
			}

			if nic.IP == "" {
				nic.IP = ip.String()
			} else {
				nic.IP += "," + ip.String()
			}
			// logger.
			// 	WithField("name", nic.Name).
			// 	WithField("ip", nic.IP).
			// 	// WithField("mask_size", nic.MaskSize).
			// 	Info("IP address found on host")

			if ip4 := ip.To4(); ip4 != nil {
				// IPv4 case
				logger.
					WithField("name", nic.Name).
					WithField("ip", ip).
					Info("IPv4 address found on host")

				s.IPVersion = 4
			} else {
				// IPv6 case
				logger.
					WithField("name", nic.Name).
					WithField("ip", ip).
					Info("IPv6 address found on host")

				s.IPVersion = 6
			}

			// ignore 127.0.0.1 || fe80:
			if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue

			}

			// otherwise map
			subnets = append(subnets, s)
			key := fmt.Sprintf("%v,%v", s.NetworkCIDR, nic.MAC)
			subnetNICMapper[key] = true
		}

		// add the NIC
		nics = append(nics, nic)
		// machine.NICS = append(machine.NICS, &nic)
	}

	// create subnets
	// sout := make([]models.Subnetwork, 0)
	err = storage.DB().
		NewInsert().
		Model(&subnets).
		On("CONFLICT DO UPDATE").
		Set("updated_at = CURRENT_TIMESTAMP").
		Scan(ctx)
	if err != nil {
		return fmt.Errorf("unable to insert new subnets: %v", err)
	}

	// fmt.Println("NICS to insert:", nics)
	// nout := make([]models.NetworkInterface, 0)
	err = storage.DB().NewInsert().
		Model(&nics).
		On("CONFLICT DO UPDATE").
		Set("mac = EXCLUDED.mac").
		Set("ip = EXCLUDED.ip").
		Set("gateway = EXCLUDED.gateway").
		Set("flags = EXCLUDED.flags").
		Set("updated_at = CURRENT_TIMESTAMP").
		Scan(ctx)

	// update nics with subnetID
	links := make([]models.NetworkInterfaceSubnet, 0)
	for _, subnet := range subnets {
		for _, nic := range nics {
			key := fmt.Sprintf("%v,%v", subnet.NetworkCIDR, nic.MAC)
			if _, ok := subnetNICMapper[key]; ok {
				link := models.NetworkInterfaceSubnet{
					NetworkInterfaceID: nic.ID,
					SubnetworkID:       subnet.ID,
				}
				links = append(links, link)
			}
		}
	}

	// insert links
	if len(links) > 0 {
		_, err = storage.DB().
			NewInsert().
			Model(&links).
			On("CONFLICT DO NOTHING").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("unable to insert network interface - subnetwork links: %v", err)
		}
	}

	return err
}

// type NetworkInterface struct {
// 	Name  string
// 	MAC   net.HardwareAddr
// 	Addrs []*net.IPNet
// }

// extractNetworks returns the address (sub-networks actually)
// attached to an interface. If keepLocalAddress is set to True
// the IP field of net.IPNet instance is set to the local IP address
// func extractNetworks(iface net.Interface, keepLocalAddress bool) []*net.IPNet {
// 	addrs, err := iface.Addrs()
// 	if err != nil {
// 		return nil
// 	}

// 	nets := make([]*net.IPNet, 0)
// 	for _, addr := range addrs {
// 		ip, ipnet, err := net.ParseCIDR(addr.String())
// 		if err != nil {
// 			continue
// 		}
// 		if keepLocalAddress {
// 			ipnet.IP = ip
// 		}
// 		nets = append(nets, ipnet)
// 	}

// 	return nets
// }

func filterInterfaces(interfaces []net.Interface) []net.Interface {
	filtered := make([]net.Interface, 0)
	for _, iface := range interfaces {
		// ignore loopback
		// if (iface.Flags & net.FlagLoopback) != 0 {
		// 	continue
		// }
		// ignore non up
		if (iface.Flags & net.FlagUp) == 0 {
			continue
		}

		// ignore docker (not anymore)
		// if strings.HasPrefix(iface.Name, "docker") {
		// 	continue
		// }
		// ignore docker bridge (not anymore)
		// if strings.HasPrefix(iface.Name, "br-") {
		// 	continue
		// }
		// ignore docker virtual interface (not anymore)
		if strings.HasPrefix(iface.Name, "veth") {
			continue
		}
		// ignore libvirt bridge
		// if strings.HasPrefix(iface.Name, "virbr") {
		// 	continue
		// }
		// ignore qemu bridge
		if strings.Contains(iface.Name, "qemu") {
			continue
		}
		filtered = append(filtered, iface)
	}
	return filtered
}

// getInterfaces returns the local relevant interfaces
func getInterfaces() ([]net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	// filter
	ifaces = filterInterfaces(ifaces)
	return ifaces, nil
}

func gateway() (*net.Interface, net.IP, error) {
	r, err := netroute.New()
	if err != nil {
		return nil, nil, err
	}
	iface, gw, _, err := r.Route(GOOGLE)
	return iface, gw, err
}

func gatewayWithSrc(hw net.HardwareAddr, src net.IP) (*net.Interface, net.IP, error) {
	r, err := netroute.New()
	if err != nil {
		return nil, nil, err
	}
	iface, gw, _, err := r.RouteWithSrc(hw, src, GOOGLE)
	return iface, gw, err
}

// func copyHardwareAddr(from net.HardwareAddr) net.HardwareAddr {
// 	dest := make(net.HardwareAddr, len(from))
// 	copy(dest, from)
// 	return dest
// }

// func GetNetworkInterfaces() ([]NetworkInterface, error) {
// 	// get the interfaces
// 	ifaces, err := getInterfaces()
// 	if err != nil {
// 		return nil, err
// 	}

// 	nics := make([]NetworkInterface, len(ifaces))
// 	// extract the networks
// 	for i, iface := range ifaces {
// 		nics[i] = NetworkInterface{
// 			Name:  iface.Name,
// 			MAC:   copyHardwareAddr(iface.HardwareAddr),
// 			Addrs: extractNetworks(iface, true),
// 		}
// 	}
// 	return nics, nil
// }

// GetInternetGateway returns the IP of the next hop while reaching the Internet
// func GetInternetGateway() net.IP {
// 	ip := net.IPv4(8, 8, 8, 8)

// 	router, err := routing.New()
// 	if err != nil {
// 		return nil
// 	}

// 	_, gw, _, err := router.Route(ip)
// 	if err != nil {
// 		return nil
// 	}

// 	return gw
// }

// GetPublicIP returns the ip of the machine hosting the agent
// when it makes outside requests (gateway public ip)
// func GetPublicIP() net.IP {
// 	// m.logger.Debugf("Joining api.ipify.org")
// 	resp, err := http.Get("https://api.ipify.org?format=txt")
// 	if err != nil {
// 		// m.logger.Error(err)
// 		return nil
// 	}

// 	// m.logger.Debugf("Parsing response body")
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		// m.logger.Error(err)
// 		return nil
// 	}
// 	// If body is not a valid textual representation of an IP address, ParseIP returns nil.
// 	return net.ParseIP(string(body))
// }
