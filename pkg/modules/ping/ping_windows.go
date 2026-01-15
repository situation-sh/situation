//go:build windows

package ping

import (
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var iphlpapi *windows.DLL

func init() {
	iphlpapi = windows.MustLoadDLL("Iphlpapi.dll")
}

// ipStatus represents the status of an ICMP echo reply
// https://learn.microsoft.com/en-us/windows/win32/api/ipexport/ns-ipexport-icmp_echo_reply
type ipStatus uint32

const (
	ipStatusSuccess             ipStatus = 0
	ipStatusBufferTooSmall      ipStatus = 11001
	ipStatusDestNetUnreachable  ipStatus = 11002
	ipStatusDestHostUnreachable ipStatus = 11003
	ipStatusDestProtUnreachable ipStatus = 11004
	ipStatusDestPortUnreachable ipStatus = 11005
	ipStatusNoResources         ipStatus = 11006
	ipStatusBadOption           ipStatus = 11007
	ipStatusHwError             ipStatus = 11008
	ipStatusPacketTooBig        ipStatus = 11009
	ipStatusReqTimedOut         ipStatus = 11010
	ipStatusBadReq              ipStatus = 11011
	ipStatusBadRoute            ipStatus = 11012
	ipStatusTTLExpiredTransit   ipStatus = 11013
	ipStatusTTLExpiredReassem   ipStatus = 11014
	ipStatusParamProblem        ipStatus = 11015
	ipStatusSourceQuench        ipStatus = 11016
	ipStatusOptionTooBig        ipStatus = 11017
	ipStatusBadDest             ipStatus = 11018
	ipStatusGeneralFailure      ipStatus = 11050
)

func (s ipStatus) Error() error {
	switch s {
	case ipStatusSuccess:
		return nil
	case ipStatusBufferTooSmall:
		return fmt.Errorf("The reply buffer was too small. ")
	case ipStatusDestNetUnreachable:
		return fmt.Errorf("The destination network is unreachable. ")
	case ipStatusDestHostUnreachable:
		return fmt.Errorf("The destination host is unreachable. ")
	case ipStatusDestProtUnreachable:
		return fmt.Errorf("The destination protocol is unreachable. ")
	case ipStatusDestPortUnreachable:
		return fmt.Errorf("The destination port is unreachable. ")
	case ipStatusNoResources:
		return fmt.Errorf("Insufficient resources are available to complete the request. ")
	case ipStatusBadOption:
		return fmt.Errorf("A bad option was specified. ")
	case ipStatusHwError:
		return fmt.Errorf("A hardware error occurred. ")
	case ipStatusPacketTooBig:
		return fmt.Errorf("The packet is too big to transmit. ")
	case ipStatusReqTimedOut:
		return fmt.Errorf("The request timed out. ")
	case ipStatusBadReq:
		return fmt.Errorf("A bad ICMP request was specified. ")
	case ipStatusBadRoute:
		return fmt.Errorf("A bad route was specified. ")
	case ipStatusTTLExpiredTransit:
		return fmt.Errorf("The time to live (TTL) expired in transit. ")
	case ipStatusTTLExpiredReassem:
		return fmt.Errorf("The time to live (TTL) expired during reassembly. ")
	case ipStatusParamProblem:
		return fmt.Errorf("A parameter problem was encountered. ")
	case ipStatusSourceQuench:
		return fmt.Errorf("A source quench message was received. ")
	case ipStatusOptionTooBig:
		return fmt.Errorf("The IP options size is too big. ")
	case ipStatusBadDest:
		return fmt.Errorf("The destination address is bad. ")
	case ipStatusGeneralFailure:
		return fmt.Errorf("A general failure occurred. ")
	default:
		return fmt.Errorf("ICMP error: %d", s)
	}
}

// ipOptionInformation represents IP options for ICMP echo request
// https://learn.microsoft.com/en-us/windows/win32/api/ipexport/ns-ipexport-ip_option_information
type ipOptionInformation struct {
	TTL         uint8
	Tos         uint8
	Flags       uint8
	OptionsSize uint8
	OptionsData uintptr
}

// icmpEchoReply represents the reply from an ICMP echo request
// https://learn.microsoft.com/en-us/windows/win32/api/ipexport/ns-ipexport-icmp_echo_reply
type icmpEchoReply struct {
	Address       uint32
	Status        ipStatus
	RoundTripTime uint32
	DataSize      uint16
	Reserved      uint16
	Data          uintptr
	Options       ipOptionInformation
}

// Ping4 sends an ICMP echo request to the target IPv4 address using Windows API.
// It calls onRecv callback if a reply is received within the timeout.
// Uses IcmpSendEcho2Ex to support specifying a source address.
func Ping4(target net.IP, source net.IP, timeout time.Duration, onRecv func(net.IP)) error {
	// Skip broadcast addresses
	if target[len(target)-1] == 0xff {
		return nil
	}

	// Convert target IP to uint32 (little-endian for Windows IPAddr)
	target4 := target.To4()
	if target4 == nil {
		return fmt.Errorf("invalid IPv4 address: %v", target)
	}
	destAddr := binary.LittleEndian.Uint32(target4)

	// Convert source IP to uint32 (0.0.0.0 if not specified)
	var srcAddr uint32 = 0
	if source != nil {
		source4 := source.To4()
		if source4 != nil {
			srcAddr = binary.LittleEndian.Uint32(source4)
		}
	}

	// Find IcmpCreateFile procedure
	icmpCreateFile, err := iphlpapi.FindProc("IcmpCreateFile")
	if err != nil {
		return fmt.Errorf("failed to find IcmpCreateFile: %w", err)
	}

	// Find IcmpSendEcho2Ex procedure
	icmpSendEcho2Ex, err := iphlpapi.FindProc("IcmpSendEcho2Ex")
	if err != nil {
		return fmt.Errorf("failed to find IcmpSendEcho2Ex: %w", err)
	}

	// Find IcmpCloseHandle procedure
	icmpCloseHandle, err := iphlpapi.FindProc("IcmpCloseHandle")
	if err != nil {
		return fmt.Errorf("failed to find IcmpCloseHandle: %w", err)
	}

	// Create ICMP handle
	handle, _, _ := icmpCreateFile.Call()
	if handle == uintptr(syscall.InvalidHandle) {
		return fmt.Errorf("failed to create ICMP handle: %w", windows.GetLastError())
	}
	defer icmpCloseHandle.Call(handle)

	// Prepare request data
	requestData := []byte("situation")
	requestSize := uint16(len(requestData))

	// Prepare reply buffer (must be large enough for ICMP_ECHO_REPLY + data + 8 bytes for ICMP header)
	replySize := uint32(unsafe.Sizeof(icmpEchoReply{})) + uint32(requestSize) + 8
	replyBuffer := make([]byte, replySize)

	// Convert timeout to milliseconds
	timeoutMs := uint32(timeout.Milliseconds())
	if timeoutMs == 0 {
		timeoutMs = 1000 // default 1 second
	}

	// Call IcmpSendEcho2Ex
	// DWORD IcmpSendEcho2Ex(
	//   HANDLE                 IcmpHandle,
	//   HANDLE                 Event,           // optional
	//   PIO_APC_ROUTINE        ApcRoutine,      // optional
	//   PVOID                  ApcContext,      // optional
	//   IPAddr                 SourceAddress,
	//   IPAddr                 DestinationAddress,
	//   LPVOID                 RequestData,
	//   WORD                   RequestSize,
	//   PIP_OPTION_INFORMATION RequestOptions,  // optional
	//   LPVOID                 ReplyBuffer,
	//   DWORD                  ReplySize,
	//   DWORD                  Timeout
	// );
	ret, _, _ := icmpSendEcho2Ex.Call(
		handle,
		0, // Event (not used for synchronous call)
		0, // ApcRoutine (not used for synchronous call)
		0, // ApcContext (not used for synchronous call)
		uintptr(srcAddr),
		uintptr(destAddr),
		uintptr(unsafe.Pointer(&requestData[0])),
		uintptr(requestSize),
		0, // RequestOptions (no options)
		uintptr(unsafe.Pointer(&replyBuffer[0])),
		uintptr(replySize),
		uintptr(timeoutMs),
	)

	// ret is the number of ICMP_ECHO_REPLY structures stored in ReplyBuffer
	if ret == 0 {
		// No reply received (timeout or error)
		// This is not an error condition for our use case
		return nil
	}

	// Parse reply
	reply := (*icmpEchoReply)(unsafe.Pointer(&replyBuffer[0]))
	if err := reply.Status.Error(); err != nil {
		return err
	}
	onRecv(target)
	return nil
}
