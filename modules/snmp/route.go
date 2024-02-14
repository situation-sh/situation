package snmp

import (
	"errors"
	"fmt"
	"net"

	"github.com/gosnmp/gosnmp"
	"github.com/situation-sh/situation/utils"
)

type snmpRoute struct {
	Destination *net.IPNet
	NextHop     net.IP
	// The type of route.
	//
	// Note that local(3) refers to a route for
	// which the next hop is the final destination; remote(4) refers
	// to a route for which the next hop is not the final destination.
	//
	// Setting this object to the value invalid(2) has the effect of
	// invalidating the corresponding entry in the ipForwardTable object.
	// That is, it effectively disassociates the destination identified
	// with said entry from the route identified with said entry.
	// It is an implementation-specific matter as to whether the agent
	// removes an invalidated entry from the table. Accordingly, management
	// stations must be prepared to receive tabular information from agents
	// that corresponds to entries not currently in use. Proper
	// interpretation of such entries requires examination of the relevant
	// ip-ForwardType object.
	//
	// 		- 1: other
	// 		- 2: invalid
	// 		- 3: local
	// 		- 4: remote
	Type int
	// The routing mechanism via which this route was learned.
	// Inclusion of values for gateway routing protocols is not
	// intended to imply that hosts should support those protocols.
	//
	//  1 - other
	//  2 - local
	//  3 - netmgmt
	//  4 - icmp
	//  5 - egp
	//  6 - ggp
	//  7 - hello
	//  8 - rip
	//  9 - is-is
	// 10 - es-is
	// 11 - ciscoIgrp
	// 12 - bbnSpfIgp
	// 13 - ospf
	// 14 - bgp
	// 15 - idpr
	Proto int
}

func (s *snmpRoute) String() string {
	main := ""
	if s.Type == 4 {
		main = "*"
	}
	return fmt.Sprintf("%s%v (%v)", main, s.Destination, s.NextHop)
}

// ipForwardTable is an obsolete snmp entry but still used on windows
// see https://cric.grenoble.cnrs.fr/Administrateurs/Outils/MIBS/?oid=1.3.6.1.2.1.4.24.2.1
// to see all the children
// return a map ifindex->routes
func ipForwardTable(g *gosnmp.GoSNMP) (map[int][]*snmpRoute, error) {
	// IpForwardDest: 		1.3.6.1.2.1.4.24.2.1.1
	// IpForwardMask: 		1.3.6.1.2.1.4.24.2.1.2
	// IpForwardNextHop: 	1.3.6.1.2.1.4.24.2.1.4
	// IpForwardIfIndex: 	1.3.6.1.2.1.4.24.2.1.5
	// IpForwardType: 		1.3.6.1.2.1.4.24.2.1.6
	// IpForwardProto: 		1.3.6.1.2.1.4.24.2.1.7
	errs := make([]error, 0)
	var oid string
	keymap := make(map[string]int)
	out := make(map[int][]*snmpRoute)
	routesMap := make(map[string]*snmpRoute)

	// ifindex
	oid = OID_IP_FORWARD_IF_INDEX
	results, err := g.BulkWalkAll(oid)
	if err != nil {
		errs = append(errs, err)
	} else {
		for _, r := range results {
			key := removePrefix(r.Name, oid)
			index, err := parseInteger(r)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			// build the key -> ifindex mapping
			keymap[key] = index
		}
	}

	// destination
	oid = OID_IP_FORWARD_DEST
	results, err = g.BulkWalkAll(oid)
	if err != nil {
		errs = append(errs, err)
	} else {
		for _, r := range results {
			key := removePrefix(r.Name, oid)
			ip, err := parseIPAddress(r)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			// init the object
			routesMap[key] = &snmpRoute{Destination: &net.IPNet{IP: ip}, Type: -1, Proto: -1}
		}
	}

	// mask
	oid = OID_IP_FORWARD_MASK
	results, err = g.BulkWalkAll(oid)
	if err != nil {
		errs = append(errs, err)
	} else {
		for _, r := range results {
			key := removePrefix(r.Name, oid)
			ipmask, err := parseIPAddress(r)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			route := routesMap[key]
			route.Destination.Mask = net.IPMask(ipmask)
		}
	}

	// nexthop
	oid = OID_IP_FORWARD_NEXT_HOP
	results, err = g.BulkWalkAll(oid)
	if err != nil {
		errs = append(errs, err)
	} else {
		for _, r := range results {
			key := removePrefix(r.Name, oid)
			hop, err := parseIPAddress(r)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			route := routesMap[key]
			route.NextHop = utils.CopyIP(hop)
		}
	}

	// type
	oid = OID_IP_FORWARD_TYPE
	results, err = g.BulkWalkAll(oid)
	if err != nil {
		errs = append(errs, err)
	} else {
		for _, r := range results {
			key := removePrefix(r.Name, oid)
			t, err := parseInteger(r)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			route := routesMap[key]
			route.Type = t
		}
	}

	// proto
	oid = OID_IP_FORWARD_PROTO
	results, err = g.BulkWalkAll(oid)
	if err != nil {
		errs = append(errs, err)
	} else {
		for _, r := range results {
			key := removePrefix(r.Name, oid)
			proto, err := parseInteger(r)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			route := routesMap[key]
			route.Proto = proto

		}
	}

	for k := range routesMap {
		index := keymap[k]
		route := routesMap[k]
		out[index] = append(out[index], route)
	}
	return out, errors.Join(errs...)
}

//	INDEX {
//		inetCidrRouteDestType,
//		inetCidrRouteDest,
//		inetCidrRoutePfxLen,
//		inetCidrRoutePolicy, {0 0} by default
//		inetCidrRouteNextHopType,
//		inetCidrRouteNextHop
//		}
//
// ::= { inetCidrRouteTable 1 }
func inetCidrTable(g *gosnmp.GoSNMP) (map[int][]*snmpRoute, error) {
	errs := make([]error, 0)
	var oid string

	keymap := make(map[string]int)
	out := make(map[int][]*snmpRoute)
	routesMap := make(map[string]*snmpRoute)

	// ifindex
	oid = OID_INET_CIDR_ROUTE_IF_INDEX
	results, err := g.BulkWalkAll(oid)
	if err != nil {
		return nil, err
	}

	for _, r := range results {
		initialKey := removePrefix(r.Name, oid)
		index, err := parseInteger(r)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		// build the key -> ifindex mapping
		keymap[initialKey] = index

		if _, ok := out[index]; !ok {
			out[index] = make([]*snmpRoute, 0)
		}

		// build the key -> ifindex mapping
		// keymap[key] = index

		// now we need to parse the key. It contains all the information.
		// inetCidrRouteDestType, -> .1.4 (meaning ipv4 and 4 bytes)
		//	inetCidrRouteDest,
		//	inetCidrRoutePfxLen,
		//	inetCidrRoutePolicy, {0 0} by default but it seems to store a dynamic number of values (size would be given by the first number)
		//	inetCidrRouteNextHopType,
		//	inetCidrRouteNextHop

		// dest
		ip, key := parseIPAddressInOid(initialKey)
		if ip == nil {
			errs = append(errs, fmt.Errorf("cannot parse ip for key: %s", key))
			continue
		}
		route := &snmpRoute{Destination: &net.IPNet{IP: ip}, Type: -1, Proto: -1}
		routesMap[initialKey] = route

		// prefixlen
		bits := 32 // ipv4
		if ip.To4() == nil {
			bits = 128 // ipv6
		}
		prefixRaw, key := splitPoint(key, 1)
		prefixlen := int(oidToUint8[prefixRaw])
		route.Destination.Mask = net.CIDRMask(prefixlen, bits)

		// policy size
		policySizeRaw, key := splitPoint(key, 1)
		policySize := int(oidToUint8[policySizeRaw])
		// skip policy
		_, key = splitPoint(key, policySize)

		// nexthop address type
		ip, key = parseIPAddressInOid(key)
		if ip == nil {
			errs = append(errs, fmt.Errorf("cannot parse ip for key: %s", key))
			continue
		}
		route.NextHop = ip
	}

	// type
	oid = OID_INET_CIDR_ROUTE_TYPE
	results, err = g.BulkWalkAll(oid)
	if err != nil {
		errs = append(errs, err)
	} else {
		for _, r := range results {
			initialKey := removePrefix(r.Name, oid)
			t, err := parseInteger(r)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			route := routesMap[initialKey]
			route.Type = t
		}
	}

	for k := range routesMap {
		index := keymap[k]
		route := routesMap[k]
		out[index] = append(out[index], route)
	}
	return out, errors.Join(errs...)
}

func ifaceRoutes(g *gosnmp.GoSNMP) (map[int][]*snmpRoute, error) {
	errs := make([]error, 0)
	// try with current interface
	routes, err := inetCidrTable(g)
	if err == nil && len(routes) == 0 {
		err = fmt.Errorf("no route found through inetCidrTable interface")
	}
	if err == nil {
		return routes, err
	}
	errs = append(errs, err)

	// legacy
	routes, err = ipForwardTable(g)
	if err == nil && len(routes) == 0 {
		err = fmt.Errorf("no route found through ipForward interface")
	}
	if err == nil {
		return routes, err
	}
	errs = append(errs, err)

	return nil, errors.Join(errs...)
}
