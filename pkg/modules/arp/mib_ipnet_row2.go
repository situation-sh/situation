//go:build windows

package arp

import (
	"encoding/binary"
	"net"
	"time"
)

// MIBIpNetRow2Size is the size in bytes of
// the MIBIpNetRow2 structure
const MIBIpNetRow2Size = 88

// SockAddrSize is the size in bytes of
// the _SOCKADDR_INET structure
const SockAddrSize = 28

//	struct sockaddr_in {
//			short   sin_family; 	// 2 bytes
//			u_short sin_port;   	// 2 bytes
//			struct  in_addr sin_addr; // 4 bytes
//			char    sin_zero[8]; 	// 8 bytes
//	};
type SockAddrIn struct {
	sinFamily uint16
	sinPort   uint16
	sinAddr   net.IP
	sinZero   []byte
}

func NewSockAddrIn(buffer []byte) SockAddrIn {
	addr := SockAddrIn{
		sinFamily: binary.LittleEndian.Uint16(buffer[:2]),
		sinPort:   binary.LittleEndian.Uint16(buffer[2:4]),
		sinAddr:   net.IP(make([]byte, 4)).To4(),
		sinZero:   make([]byte, 8),
	}
	copy(addr.sinAddr, buffer[4:8])
	copy(addr.sinZero, buffer[8:16])
	return addr
}

func (s SockAddrIn) Family() uint16 {
	return s.sinFamily
}

func (s SockAddrIn) Addr() net.IP {
	return s.sinAddr.To4()
}

//	struct sockaddr_in6 {
//			short   sin6_family; 	// 2 bytes
//			u_short sin6_port; 		// 2 bytes
//			u_long  sin6_flowinfo; 	// 4 bytes
//			struct  in6_addr sin6_addr; // 16 bytes
//			u_long  sin6_scope_id;	// 4 bytes
//	}; 44 bytes
type SockAddrIn6 struct {
	sin6Family   uint16
	sin6Port     uint16
	sin6FlowInfo uint32
	sin6Addr     net.IP
	sin6ScopeId  uint32
}

func NewSockAddrIn6(buffer []byte) SockAddrIn6 {
	addr := SockAddrIn6{
		sin6Family:   binary.LittleEndian.Uint16(buffer[:2]),
		sin6Port:     binary.LittleEndian.Uint16(buffer[2:4]),
		sin6FlowInfo: binary.LittleEndian.Uint32(buffer[4:8]),
		sin6Addr:     net.IP(make([]byte, 16)).To16(),
		sin6ScopeId:  binary.LittleEndian.Uint32(buffer[24:28]),
	}
	copy(addr.sin6Addr, buffer[8:24])
	return addr
}

func (s SockAddrIn6) Family() uint16 {
	return s.sin6Family
}

func (s SockAddrIn6) Addr() net.IP {
	return s.sin6Addr.To16()
}

//	typedef union _SOCKADDR_INET {
//			SOCKADDR_IN    Ipv4; // 16 bytes
//			SOCKADDR_IN6   Ipv6; // 28 bytes
//			ADDRESS_FAMILY si_family; // 2 bytes (u_short)
//	} SOCKADDR_INET, *PSOCKADDR_INET; // 28 bytes
type SockAddr interface {
	Family() uint16
	Addr() net.IP
}

func parseSockAddr(buffer []byte) SockAddr {
	sockType := binary.LittleEndian.Uint16(buffer[:2])
	switch sockType {
	case 2: // IPv4
		return NewSockAddrIn(buffer[:SockAddrSize])
	case 23: // IPv6
		return NewSockAddrIn6(buffer[:SockAddrSize])
	default:
		return nil
	}
}

func parsePhysicalAddress(buffer []byte, physicalAddressLength uint32) net.HardwareAddr {
	pa := make(net.HardwareAddr, physicalAddressLength)
	copy(pa, buffer[:physicalAddressLength])
	return pa
}

// MIBIpNetRow2 is a more golang version of
// the raw MIB_IPNET_ROW2 structure
type MIBIpNetRow2 struct {
	address               SockAddr
	interfaceIndex        uint32 // The local index value for the network interface associated with this IP address. This index value may change when a network adapter is disabled and then enabled, or under other circumstances, and should not be considered persistent.
	interfaceLuid         uint64 // The locally unique identifier (LUID) for the network interface associated with this IP address.
	physicalAddress       net.HardwareAddr
	physicalAddressLength uint32
	state                 uint32
	flags                 uint32
	reachabilityTime      time.Duration
}

// MAC returns a copy of the MAC address
func (r MIBIpNetRow2) MAC() net.HardwareAddr {
	mac := make(net.HardwareAddr, r.physicalAddressLength)
	copy(mac, r.physicalAddress)
	return mac
}

// IP returns a copy of the IP address
func (r MIBIpNetRow2) IP() net.IP {
	length := len(r.address.Addr())
	ip := make(net.IP, length)
	copy(ip, r.address.Addr())
	return ip
}

func (r MIBIpNetRow2) ToARPEntry() ARPEntry {
	return ARPEntry{
		Family:         int(r.address.Family()),
		InterfaceIndex: int(r.interfaceIndex),
		MAC:            r.MAC(),
		IP:             r.IP(),
		State:          WindowsState(int(r.state)),
	}
}

// rawMIBIpNetRow2 mirrors the MIB_IPNET_ROW2 structure detailed at
// https://docs.microsoft.com/en-us/windows/win32/api/netioapi/ns-netioapi-mib_ipnet_row2
//
//	typedef struct _MIB_IPNET_ROW2 {
//			SOCKADDR_INET     Address; // 28 bytes
//			NET_IFINDEX       InterfaceIndex; // 4 bytes (ulong)
//			NET_LUID          InterfaceLuid; // 8 bytes ?
//			UCHAR             PhysicalAddress[IF_MAX_PHYS_ADDRESS_LENGTH]; // 32 bytes
//			ULONG             PhysicalAddressLength; // 4 bytes
//			NL_NEIGHBOR_STATE State; // 4 bytes (enum)
//			union {
//		  		struct {
//					BOOLEAN IsRouter : 1;
//					BOOLEAN IsUnreachable : 1;
//		  		};
//		  		UCHAR Flags;
//			}; // 4 bytes (it looks like 1 byte but there are 4 in practice)
//			union {
//		  		ULONG LastReachable;
//		  		ULONG LastUnreachable;
//			} ReachabilityTime; // 4 bytes
//	} MIB_IPNET_ROW2, *PMIB_IPNET_ROW2;
//
// 88 bytes actually
type rawMIBIpNetRow2 struct {
	address               [28]byte
	interfaceIndex        uint32 // The local index value for the network interface associated with this IP address. This index value may change when a network adapter is disabled and then enabled, or under other circumstances, and should not be considered persistent.
	interfaceLuid         uint64 // The locally unique identifier (LUID) for the network interface associated with this IP address.
	physicalAddress       [32]byte
	physicalAddressLength uint32
	state                 uint32
	flags                 uint32
	reachabilityTime      uint32
}

func (r rawMIBIpNetRow2) Parse() MIBIpNetRow2 {
	return MIBIpNetRow2{
		address:               parseSockAddr(r.address[:]),
		interfaceIndex:        r.interfaceIndex,
		interfaceLuid:         r.interfaceLuid,
		physicalAddress:       parsePhysicalAddress(r.physicalAddress[:], r.physicalAddressLength),
		physicalAddressLength: r.physicalAddressLength,
		state:                 r.state,
		flags:                 r.flags,
		reachabilityTime:      time.Duration(r.reachabilityTime * uint32(time.Millisecond)),
	}
}
