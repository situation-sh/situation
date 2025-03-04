//go:build windows
// +build windows

package arp

const anySize = 1 << 16

// MIBIpNetTable2 is a "golang version" of the
// raw structure MIB_IPNET_TABLE2
type MIBIpNetTable2 []MIBIpNetRow2

// rawMIBIpNetTable2 mirrors the MIB_IPNET_TABLE2 structure
// see https://docs.microsoft.com/en-us/windows/win32/api/netioapi/ns-netioapi-mib_ipnet_table2
//
//	typedef struct _MIB_IPNET_TABLE2 {
//	  	ULONG          NumEntries;
//	  	MIB_IPNET_ROW2 Table[ANY_SIZE];
//	} MIB_IPNET_TABLE2, *PMIB_IPNET_TABLE2;
//
// In theory NumEntries must be uint32 but doc says:
// Note that the returned MIB_IPNET_TABLE2 structure pointed to by the Table
// parameter may contain padding for alignment between the NumEntries member
// and the first MIB_IPNET_ROW2 array entry in the Table member of the
// MIB_IPNET_TABLE2 structure. Padding for alignment may also be present
// between the MIB_IPNET_ROW2 array entries. Any access to a MIB_IPNET_ROW2
// array entry should assume padding may exist.
type rawMIBIpNetTable2 struct {
	// numEntries uint32
	// padding    uint32 // set empirically to 4 bytes
	numEntries uint64
	table      [anySize]rawMIBIpNetRow2
}

// parse copy the dll allocated memory into the
// more golang structure MIBIpNetTable2
func (r *rawMIBIpNetTable2) parse() MIBIpNetTable2 {
	t := make([]MIBIpNetRow2, r.numEntries)
	// parse each row
	for i := 0; i < int(r.numEntries); i++ {
		t[i] = r.table[i].Parse()
	}
	return t
}
