package arp

import (
	"fmt"
	"net"
	"testing"
)

func TestGetARPTable(t *testing.T) {
	table, err := GetARPTable()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v\n", table)
}

func TestFilterARPTableByNetwork(t *testing.T) {
	table, err := GetARPTable()
	if err != nil {
		t.Error(err)
	}
	if len(table) <= 0 {
		t.Logf("table has no entry")
		return
	}
	network := net.IPNet{IP: table[0].IP, Mask: table[0].IP.DefaultMask()}
	out := FilterARPTableByNetwork(table, &network)
	if len(out) <= 0 {
		t.Errorf("the filter must return at least one element")
	}
}

func TestState(t *testing.T) {
	for i := 0; i <= 10; i++ {
		if i <= 6 && WindowsState(i) == Unknown {
			t.Errorf("bad state associated with code %d (it must not be Unknown)", i)
		}
		if i > 6 && WindowsState(i) != Unknown {
			t.Errorf("bad state associated with code %d (it must be Unknown, not %v)", i, WindowsState(i))
		}
	}
	for _, i := range []int{0x00, 0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80} {
		if LinuxState(i) == Unknown {
			t.Errorf("bad state associated with code %d (it must not be Unknown)", i)
		}
	}
}
