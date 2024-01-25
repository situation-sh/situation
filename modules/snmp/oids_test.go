package snmp

import (
	"net"
	"testing"
)

func TestParseIPAddressInOid(t *testing.T) {
	oid := "2.16.254.128.0.0.0.0.0.0.147.108.137.216.161.31.176.130.x.y.z"
	ip, remain := parseIPAddressInOid(oid)
	if remain != "x.y.z" {
		t.Errorf("bad remain: %s", remain)
	}
	ip6 := net.ParseIP("fe80::936c:89d8:a11f:b082")

	if !ip.Equal(ip6) {
		t.Errorf("bad ip: %v", ip)
	}
}
