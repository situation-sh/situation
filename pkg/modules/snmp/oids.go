package snmp

import (
	"fmt"
	"net"
	"strings"

	"github.com/gosnmp/gosnmp"
)

const (
	OID_BASE       = "1.3.6.1"
	OID_MGMT       = OID_BASE + ".2"
	OID_MIB2       = OID_MGMT + ".1"
	OID_SYSTEM     = OID_MIB2 + ".1"
	OID_INTERFACES = OID_MIB2 + ".2"
	OID_IP         = OID_MIB2 + ".4"
	OID_TCP        = OID_MIB2 + ".6"

	OID_INTERFACES_NUMBER          = OID_INTERFACES + ".1.0"
	OID_INTERFACES_IF_TABLE        = OID_INTERFACES + ".2.1"
	OID_INTERFACES_IF_INDEX        = OID_INTERFACES_IF_TABLE + ".1"
	OID_INTERFACES_IF_TYPE         = OID_INTERFACES_IF_TABLE + ".3"
	OID_INTERFACES_IF_PHYS_ADDRESS = OID_INTERFACES_IF_TABLE + ".6"
	OID_INTERFACES_IF_NAME         = OID_MIB2 + ".31.1.1.1.1"

	OID_IP_ADDR_ENTRY          = OID_IP + ".20.1"
	OID_IP_ADDR_ENTRY_ADDR     = OID_IP_ADDR_ENTRY + ".1"
	OID_IP_ADDR_ENTRY_IF_INDEX = OID_IP_ADDR_ENTRY + ".2"
	OID_IP_ADDR_ENTRY_NET_MASK = OID_IP_ADDR_ENTRY + ".3"

	OID_IP_ADDRESS_ENTRY    = OID_IP + ".34.1"
	OID_IP_ADDRESS_IF_INDEX = OID_IP_ADDRESS_ENTRY + ".3"

	OID_IP_ADDRESS_PREFIX_ENTRY  = OID_IP + ".32.1"
	OID_IP_ADDRESS_PREFIX_ORIGIN = OID_IP_ADDRESS_PREFIX_ENTRY + ".5"

	OID_IP_FORWARD_ENTRY    = OID_IP + ".24.2.1"
	OID_IP_FORWARD_DEST     = OID_IP_FORWARD_ENTRY + ".1"
	OID_IP_FORWARD_MASK     = OID_IP_FORWARD_ENTRY + ".2"
	OID_IP_FORWARD_NEXT_HOP = OID_IP_FORWARD_ENTRY + ".4"
	OID_IP_FORWARD_IF_INDEX = OID_IP_FORWARD_ENTRY + ".5"
	OID_IP_FORWARD_TYPE     = OID_IP_FORWARD_ENTRY + ".6"
	OID_IP_FORWARD_PROTO    = OID_IP_FORWARD_ENTRY + ".7"

	OID_INET_CIDR_ROUTE_ENTRY    = OID_IP + ".24.7.1"
	OID_INET_CIDR_ROUTE_IF_INDEX = OID_INET_CIDR_ROUTE_ENTRY + ".7"
	OID_INET_CIDR_ROUTE_TYPE     = OID_INET_CIDR_ROUTE_ENTRY + ".8"
)

func removePrefix(fullOid string, prefixOid string) string {
	if fullOid[0] == '.' && prefixOid[0] != '.' {
		prefixOid = "." + prefixOid
	}
	if prefixOid[len(prefixOid)-1] != '.' {
		prefixOid += "."
	}
	return strings.ReplaceAll(fullOid, prefixOid, "")
}

func parseInteger(r gosnmp.SnmpPDU) (int, error) {
	v, ok := r.Value.(int)
	if !ok {
		return 0, fmt.Errorf("cannot cast %v into int (Go type: %T, Object type: %v)",
			r.Value, r.Value, r.Type)
	}
	return v, nil
}

func parseOctetString(r gosnmp.SnmpPDU) ([]byte, error) {
	v, ok := r.Value.([]uint8)
	if !ok {
		return nil, fmt.Errorf("cannot cast %v into []uint8 (Go type: %T, Object type: %v)",
			r.Value, r.Value, r.Type)
	}
	return []byte(v), nil
}

func parseIPAddress(r gosnmp.SnmpPDU) (net.IP, error) {
	v, ok := r.Value.(string)
	if !ok {
		return nil, fmt.Errorf("cannot cast %v into string (Go type: %T, Object type: %v)",
			r.Value, r.Value, r.Type)
	}
	x := net.ParseIP(v)
	if x == nil {
		return nil, fmt.Errorf("cannot turn %v (%T) into net.IP", v, v)
	}
	return x.To4(), nil
}

func splitPoint(oid string, count int) (string, string) {
	if count <= 0 {
		return oid, ""
	}
	c := 0
	for i, r := range oid {
		if r == '.' {
			c += 1
			if c == count {
				return oid[:i], oid[i+1:]
			}
		}
	}
	return oid, ""
}

func parseIPAddressInOid(oid string) (net.IP, string) {
	t, key := splitPoint(oid, 2)
	if t == "1.4" { // ipv4
		ipRaw, remain := splitPoint(key, 4)
		return net.ParseIP(ipRaw), remain
	} else if t == "2.16" || t == "4.16" { // ipv6 / ipv6z
		ipRaw, remain := splitPoint(key, 16)

		r := ipRaw // 254.128.0.0.0.0.0.0.147.108.137.216.161.31.176.130
		x := ""
		buffer := make([]byte, 16)
		for i := 0; i < 8; i++ {
			x, r = splitPoint(r, 2) // x: 254.128
			chunks := strings.Split(x, ".")
			if len(chunks) == 2 {
				buffer[2*i] = oidToUint8[chunks[0]]   // 254
				buffer[2*i+1] = oidToUint8[chunks[1]] // 128
			} else {
				fmt.Println("BAD CHUNK:", chunks)
			}
		}
		return net.IP(buffer), remain
	} else {
		// ignore fmt.Println("BAD Type:", t)
		return nil, oid
	}
}
