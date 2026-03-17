package snmp

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/situation-sh/situation/pkg/models"
)

type snmpNetwork struct {
	net.IPNet
	PrefixOrigin int
}

func (n *snmpNetwork) String() string {
	return fmt.Sprintf("%s(%d)", n.IPNet.String(), n.PrefixOrigin)
}

// SNMPInterface represents a network interface discovered via SNMP.
type SNMPInterface struct {
	Index  int
	Name   string
	MAC    net.HardwareAddr
	Type   int // see https://www.iana.org/assignments/ianaiftype-mib/ianaiftype-mib for details. Basically we have 6: ethernet, 24: loopback
	IP     []*snmpNetwork
	Routes []*snmpRoute
}

// Gateway outputs the more generic IPv4 nexthop of this interface.
func (s *SNMPInterface) Gateway() string {

	for _, route := range s.Routes {
		// only IPv4 for this function
		if route.Destination.IP.To4() == nil {
			continue
		}

		if route.Type == 4 { // remote route
			return route.NextHop.String()
		}
	}

	return ""
}

// ToNetworkInterface converts an SNMPInterface to a models.NetworkInterface.
func (s *SNMPInterface) ToNetworkInterface() *models.NetworkInterface {
	nic := models.NetworkInterface{
		Name:    s.Name,
		MAC:     s.MAC.String(),
		IP:      make([]string, 0),
		Gateway: s.Gateway(),
	}

	// we take the first IPv4 and first IPv6
	for _, n := range s.IP {
		nic.IP = append(nic.IP, n.IP.String())
	}
	return &nic
}

func interfaceCount(g *gosnmp.GoSNMP) (int, error) {
	pkt, err := g.Get([]string{OID_INTERFACES_NUMBER})
	if err != nil {
		return -1, err
	}
	if len(pkt.Variables) == 0 {
		return 0, nil
	}
	pdu := pkt.Variables[0]
	return parseInteger(pdu)
}

func getInterface(g *gosnmp.GoSNMP, index int) (*SNMPInterface, error) {
	pkt, err := g.Get([]string{
		fmt.Sprintf("%s.%d", OID_INTERFACES_IF_INDEX, index),        // index
		fmt.Sprintf("%s.%d", OID_INTERFACES_IF_NAME, index),         // name
		fmt.Sprintf("%s.%d", OID_INTERFACES_IF_PHYS_ADDRESS, index), // mac
		fmt.Sprintf("%s.%d", OID_INTERFACES_IF_TYPE, index),         // type
	})
	if err != nil {
		return nil, err
	}
	if len(pkt.Variables) < 4 {
		return nil, fmt.Errorf("bad number of pdu: %v", pkt.Variables)
	}

	iface := SNMPInterface{}
	if i, err := parseInteger(pkt.Variables[0]); err == nil {
		iface.Index = i
	}
	if name, err := parseOctetString(pkt.Variables[1]); err == nil {
		iface.Name = string(name)
	}
	if mac, err := parseOctetString(pkt.Variables[2]); err == nil {
		iface.MAC = net.HardwareAddr(mac)
	}
	if t, err := parseInteger(pkt.Variables[3]); err == nil {
		iface.Type = t
	}

	return &iface, nil
}

// GetAllInterfaces retrieves all network interfaces from the SNMP target.
func GetAllInterfaces(g *gosnmp.GoSNMP) ([]*SNMPInterface, error) {
	ifaces := make([]*SNMPInterface, 0)
	errs := make([]error, 0)

	n, err := interfaceCount(g)
	if err != nil {
		return nil, err
	}

	for i := 1; i <= n; i++ {
		iface, err := getInterface(g, i)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ifaces = append(ifaces, iface)
	}

	// populate with networks
	ifn, err := ifaceNetworks(g)
	if err == nil {
		for _, iface := range ifaces {
			if nets, ok := ifn[iface.Index]; ok {
				iface.IP = nets
			}
		}
	} else {
		errs = append(errs, err)
	}

	// populate iface with routes
	ifr, err := ifaceRoutes(g)
	if err == nil {
		for _, iface := range ifaces {
			if routes, ok := ifr[iface.Index]; ok {
				iface.Routes = append(iface.Routes, routes...)
			}
		}
	} else {
		errs = append(errs, err)
	}

	return ifaces, errors.Join(errs...)
}

func GetSystemName(g *gosnmp.GoSNMP) (string, error) {
	pkt, err := g.Get([]string{OID_SYSTEM_NAME})
	if err != nil {
		return "", err
	}
	if len(pkt.Variables) == 0 {
		return "", fmt.Errorf("no system name found")
	}
	name, err := parseOctetString(pkt.Variables[0])
	if err != nil {
		return "", err
	}
	return string(name), nil
}

func GetSystemUptime(g *gosnmp.GoSNMP) (time.Duration, error) {
	pkt, err := g.Get([]string{OID_SYSTEM_UPTIME})
	if err != nil {
		return 0, err
	}
	if len(pkt.Variables) == 0 {
		return 0, fmt.Errorf("no system uptime found")
	}
	ticks, err := parseUint32(pkt.Variables[0])
	if err != nil {
		return 0, err
	}
	// each "tick" = 1/100th of a second (10 milliseconds).
	return time.Duration(ticks) * 10 * time.Millisecond, nil
}

func GetDescription(g *gosnmp.GoSNMP) (string, error) {
	pkt, err := g.Get([]string{OID_SYSTEM_DESCRIPTION})
	if err != nil {
		return "", err
	}
	if len(pkt.Variables) == 0 {
		return "", fmt.Errorf("no system description found")
	}
	description, err := parseOctetString(pkt.Variables[0])
	if err != nil {
		return "", err
	}
	return string(description), nil

	// ld := strings.ToLower(string(description))
	// if strings.Contains(ld, "linux") {
	// 	platform = "linux"
	// }
	// if strings.Contains(ld, "windows") {
	// 	platform = "windows"
	// }
	// if strings.Contains(ld, "debian") {
	// 	distribution = "debian"
	// }
	// if strings.Contains(ld, "ubuntu") {
	// 	distribution = "ubuntu"
	// }
	// return platform, distribution, nil
}

func populateMachineFromDescription(machine *models.Machine, description string) bool {
	toUpdate := false
	ld := strings.ToLower(description)
	// platform
	if strings.Contains(ld, "linux") {
		machine.Platform = "linux"
		toUpdate = true
	}
	if strings.Contains(ld, "windows") {
		machine.Platform = "windows"
		toUpdate = true
	}
	// distribution
	if strings.Contains(ld, "debian") {
		machine.Distribution = "debian"
		machine.DistributionFamily = "debian"
		toUpdate = true
	}
	if strings.Contains(ld, "ubuntu") {
		machine.Distribution = "ubuntu"
		machine.DistributionFamily = "debian"
		toUpdate = true
	}
	// arch
	if strings.Contains(ld, "x86_64") {
		machine.Arch = "x86_64"
		toUpdate = true
	}
	if strings.Contains(ld, "aarch64") {
		machine.Arch = "aarch64"
		toUpdate = true
	}
	if strings.Contains(ld, "arm64") {
		machine.Arch = "arm64"
		toUpdate = true
	}
	if strings.Contains(ld, "i386") {
		machine.Arch = "i386"
		toUpdate = true
	}
	return toUpdate
}

func PopulateSystem(g *gosnmp.GoSNMP, machine *models.Machine) (bool, error) {
	updated := false
	errs := make([]error, 0)

	if name, err := GetSystemName(g); err == nil {
		machine.Hostname = name
		updated = true
	} else {
		errs = append(errs, err)
	}

	if uptime, err := GetSystemUptime(g); err == nil {
		machine.Uptime = uptime
		updated = true
	} else {
		errs = append(errs, err)
	}

	if description, err := GetDescription(g); err == nil {
		if populateMachineFromDescription(machine, description) {
			updated = true
		}
	} else {
		errs = append(errs, err)
	}

	return updated, errors.Join(errs...)
}
