package snmp

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/gosnmp/gosnmp"
)

// there are two ways to get ip addresses
// - using 1.3.6.1.2.1.4.20.1 ipAddrTable
// - using 1.3.6.1.2.1.4.32.1 ipAddressPrefixTable
// - using 1.3.6.1.2.1.4.34.1 ipAddressEntry (but no prefixlen)

// ipAddressPrefix is not enough to get the righ IP address. It gives only the
// subnetwork...
func ipAddressPrefix(g *gosnmp.GoSNMP) (map[int][]*snmpNetwork, error) {
	errs := make([]error, 0)
	out := make(map[int][]*snmpNetwork)

	oid := OID_IP_ADDRESS_PREFIX_ORIGIN
	results, err := g.BulkWalkAll(oid)
	if err != nil {
		return nil, err
	}

	for _, r := range results {
		po, err := parseInteger(r)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		remain := removePrefix(r.Name, oid)
		rawIfIndex, remain := splitPoint(remain, 1)
		ifIndex, err := strconv.Atoi(rawIfIndex)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		ip, remain := parseIPAddressInOid(remain)
		if ip == nil {
			errs = append(errs, fmt.Errorf("cannot parse IP in oid: %s", remain))
			continue
		}

		// last is prefixlen
		prefixLen, err := strconv.Atoi(remain)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		network := snmpNetwork{
			net.IPNet{
				IP:   ip,
				Mask: net.CIDRMask(prefixLen, len(ip)*8),
			},
			po,
		}

		if _, ok := out[ifIndex]; !ok {
			out[ifIndex] = make([]*snmpNetwork, 0)
		}
		out[ifIndex] = append(out[ifIndex], &network)
	}

	// For that we have to query OID_IP_ADDRESS_ENTRY
	// after to update the IP of found networks
	oid = OID_IP_ADDRESS_IF_INDEX
	results, err = g.BulkWalkAll(oid)
	if err != nil {
		return nil, err
	} else {
		for _, r := range results {
			index, err := parseInteger(r)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if _, ok := out[index]; !ok {
				errs = append(errs, fmt.Errorf("no interface found with index %d", index))
				continue
			}

			key := removePrefix(r.Name, oid)
			ip, key := parseIPAddressInOid(key)
			// if the ip is neither ipv4 nor ipv6 it returns nil
			// Example: ipv6z (see https://www.circitor.fr/Mibs/Html/I/INET-ADDRESS-MIB.php#InetAddressIPv6z)
			if ip == nil {
				continue
			}

			for _, n := range out[index] {
				if n.Contains(ip) {
					n.IP = ip
				}
			}
		}
	}

	return out, errors.Join(errs...)
}

func ipAddr(g *gosnmp.GoSNMP) (map[int][]*snmpNetwork, error) {
	errs := make([]error, 0)
	out := make(map[int][]*snmpNetwork)
	mapper := make(map[string]*snmpNetwork)

	oid := OID_IP_ADDR_ENTRY_ADDR
	results, err := g.BulkWalkAll(oid)
	if err != nil {
		return nil, err
	}
	for _, r := range results {

		key := removePrefix(r.Name, oid)

		ip, err := parseIPAddress(r)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		// create structure and populate mapper
		mapper[key] = &snmpNetwork{net.IPNet{IP: ip}, -1}
	}

	// mask
	oid = OID_IP_ADDR_ENTRY_NET_MASK
	results, err = g.BulkWalkAll(oid)
	if err != nil {
		return nil, err
	}
	for _, r := range results {

		key := removePrefix(r.Name, oid)

		mask, err := parseIPAddress(r)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		// update mapper
		mapper[key].IPNet.Mask = net.IPMask(mask)
	}

	// if index
	oid = OID_IP_ADDR_ENTRY_IF_INDEX
	results, err = g.BulkWalkAll(oid)
	if err != nil {
		return nil, err
	}
	for _, r := range results {

		key := removePrefix(r.Name, oid)

		ifIndex, err := parseInteger(r)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		// populate out structure
		if _, ok := out[ifIndex]; !ok {
			out[ifIndex] = make([]*snmpNetwork, 0)
		}
		out[ifIndex] = append(out[ifIndex], mapper[key])
	}

	return out, errors.Join(errs...)
}

func ifaceNetworks(g *gosnmp.GoSNMP) (map[int][]*snmpNetwork, error) {
	errs := make([]error, 0)

	// legacy
	ips, err := ipAddr(g)
	if err == nil && len(ips) == 0 {
		err = fmt.Errorf("no IP found through ipAddr interface")
	}
	if err == nil {
		return ips, err
	}
	errs = append(errs, err)

	// try with current interface
	ips, err = ipAddressPrefix(g)
	if err == nil && len(ips) == 0 {
		err = fmt.Errorf("no IP found through ipAddressPrefix interface")
	}
	if err == nil {
		return ips, err
	}
	errs = append(errs, err)

	return nil, errors.Join(errs...)
}
