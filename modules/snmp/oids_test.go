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

func TestSplitPoint(t *testing.T) {
	he := "2.16.254"
	ta := "128.0.0.0.0.0.0.147.108"
	oid := he + "." + ta
	head, tail := splitPoint(oid, 3)
	if head != he {
		t.Errorf("%s != %s", head, he)
	}
	if tail != ta {
		t.Errorf("%s != %s", tail, ta)
	}

	if head, tail := splitPoint(oid, 0); head != oid || tail != "" {
		t.Errorf("%s != %s or %s != %s", head, oid, tail, "")
	}

	if head, tail := splitPoint(oid, 10000); head != oid || tail != "" {
		t.Errorf("%s != %s or %s != %s", head, oid, tail, "")
	}
}

func TestRemovePrefix(t *testing.T) {
	he := "2.16.254"
	ta := "128.0.0.0.0.0.0.147.108"
	oid := he + "." + ta
	if out := removePrefix(oid, he); out != ta {
		t.Errorf("%s != %s", out, ta)
	}
}
