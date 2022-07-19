package arp

import (
	"fmt"
	"testing"
)

func TestGetARPTable(t *testing.T) {
	table, err := GetARPTable()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v\n", table)
}
