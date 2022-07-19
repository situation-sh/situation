package modules

import (
	"fmt"
	"net"
	"strings"

	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/store"
)

func init() {
	RegisterModule(&HostNetworkModule{})
}

// Module definition ---------------------------------------------------------

type HostNetworkModule struct{}

func (m *HostNetworkModule) Name() string {
	return "host-network"
}

func (m *HostNetworkModule) Dependencies() []string {
	return []string{"host-basic"}
}

func (m *HostNetworkModule) Run() error {
	logger := GetLogger(m)
	machine := store.GetHost()
	if machine == nil {
		return fmt.Errorf("cannot retrieve host machine")
	}

	ifaces, err := getInterfaces()
	if err != nil {
		return fmt.Errorf("error while getting network interfaces: %v", err)
	}

	// check previous network interfaces
	if len(machine.NICS) > 0 {
		return fmt.Errorf("some network interfaces have already been discovered: %+v",
			machine.NICS)
	}

	// create nics
	for _, iface := range ifaces {
		nic := models.NetworkInterface{}
		// name
		nic.Name = iface.Name
		// mac
		nic.MAC = iface.HardwareAddr

		// logging
		logger.WithField("name", nic.Name).WithField("mac", nic.MAC).Info("Network Interface found on host")

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

			if ip4 := ip.To4(); ip4 != nil {
				// IPv4 case
				// ignore if IP has already been assigned
				if nic.IP == nil {
					nic.IP = ip4
					nic.MaskSize, _ = ipnet.Mask.Size()
					logger.WithField("name", nic.Name).
						WithField("ip", nic.IP).
						WithField("mask_size", nic.MaskSize).
						Info("IPv4 address found on host")
				}
			} else {
				// IPv6 case
				// ignore if IP6 has already been assigned
				if nic.IP6 == nil {
					nic.IP6 = ip
					nic.Mask6Size, _ = ipnet.Mask.Size()
					logger.WithField("name", nic.Name).
						WithField("ip6", nic.IP6).
						WithField("mask6_size", nic.Mask6Size).
						Info("IPv6 address found on host")
				}
			}
		}

		// add the NIC
		machine.NICS = append(machine.NICS, &nic)
	}
	return nil
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
		if (iface.Flags & net.FlagLoopback) != 0 {
			continue
		}
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
// 	// logger.Debugf("Joining api.ipify.org")
// 	resp, err := http.Get("https://api.ipify.org?format=txt")
// 	if err != nil {
// 		// logger.Error(err)
// 		return nil
// 	}

// 	// logger.Debugf("Parsing response body")
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		// logger.Error(err)
// 		return nil
// 	}
// 	// If body is not a valid textual representation of an IP address, ParseIP returns nil.
// 	return net.ParseIP(string(body))
// }
