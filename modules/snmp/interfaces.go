package snmp

import (
	"errors"
	"fmt"
	"net"

	"github.com/gosnmp/gosnmp"
	"github.com/situation-sh/situation/models"
)

type snmpNetwork struct {
	net.IPNet
	PrefixOrigin int
}

func (n *snmpNetwork) String() string {
	return fmt.Sprintf("%s(%d)", n.IPNet.String(), n.PrefixOrigin)
}

type snmpInterface struct {
	Index  int
	Name   string
	MAC    net.HardwareAddr
	Type   int // see https://www.iana.org/assignments/ianaiftype-mib/ianaiftype-mib for details. Basically we have 6: ethernet, 24: loopback
	IP     []*snmpNetwork
	Routes []*snmpRoute
}

// gateway outputs the more generic IPv4 nexthop
func (s *snmpInterface) gateway() net.IP {

	for _, route := range s.Routes {
		// only IPv4 for this function
		if route.Destination.IP.To4() == nil {
			continue
		}

		if route.Type == 4 { // remote route
			return route.NextHop
		}
	}

	return nil
}

func (s *snmpInterface) toNetworkInterface() *models.NetworkInterface {
	nic := models.NetworkInterface{
		Name:    s.Name,
		MAC:     s.MAC,
		IP:      nil,
		IP6:     nil,
		Gateway: s.gateway(),
	}

	// we take the first IPv4 and first IPv6
	for _, n := range s.IP {
		isIPv4 := (n.IP.To4() != nil)
		if isIPv4 && nic.IP == nil { // IPv4
			nic.IP = n.IP
			nic.MaskSize, _ = n.Mask.Size()
		} else if !isIPv4 && nic.IP6 == nil { // IPv6
			nic.IP6 = n.IP
			nic.Mask6Size, _ = n.Mask.Size()
		}
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

func getInterface(g *gosnmp.GoSNMP, index int) (*snmpInterface, error) {
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

	iface := snmpInterface{}
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

func getAllInterfaces(g *gosnmp.GoSNMP) ([]*snmpInterface, error) {
	ifaces := make([]*snmpInterface, 0)
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
