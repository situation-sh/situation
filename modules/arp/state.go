package arp

// ARPEntryState details the statsu of the
// ARP record. The following descriptions
// have been taken from  https://docs.microsoft.com/en-us/windows/win32/api/netioapi/nf-netioapi-getipnettable2
type ARPEntryState string

const (
	// The IP address is unreachable.
	Unreachable ARPEntryState = "unreachable"

	// Address resolution is in progress and the link-layer address of
	// the neighbor has not yet been determined. Specifically for IPv6,
	// a Neighbor Solicitation has been sent to the solicited-node
	// multicast IP address of the target, but the corresponding neighbor
	// advertisement has not yet been received.
	Incomplete ARPEntryState = "incomplete"

	// The neighbor is no longer known to be reachable, and probes are
	// being sent to verify reachability. For IPv6, a reachability
	// confirmation is actively being sought by retransmitting unicast
	// Neighbor Solicitation probes at regular intervals until a
	// reachability confirmation is received.
	Probe ARPEntryState = "probe"

	// The neighbor is no longer known to be reachable, and traffic has
	// recently been sent to the neighbor. Rather than probe the neighbor
	// immediately, however, delay sending probes for a short while in
	// order to give upper layer protocols a chance to provide reachability
	// confirmation. For IPv6, more time has elapsed than is specified in
	// the ReachabilityTime.ReachableTime member since the last positive
	// confirmation was received that the forward path was functioning
	// properly and a packet was sent. If no reachability confirmation
	// is received within a period of time (used to delay the first probe)
	// of entering the NlnsDelay state, then a neighbor solicitation is
	// sent and the State member is changed to NlnsProbe.
	Delay ARPEntryState = "delay"

	// The neighbor is no longer known to be reachable but until traffic
	// is sent to the neighbor, no attempt should be made to verify its
	// reachability. For IPv6, more time has elapsed than is specified in
	// the ReachabilityTime.ReachableTime member since the last positive
	// confirmation was received that the forward path was functioning
	// properly. While the State is NlnsStale, no action takes place until
	// a packet is sent.
	// The Stale state is entered upon receiving an unsolicited neighbor
	// discovery message that updates the cached IP address. Receipt of
	// such a message does not confirm reachability, and entering the Stale
	// state insures reachability is verified quickly if the entry is
	// actually being used. However, reachability is not actually verified
	// until the entry is actually used.
	Stale ARPEntryState = "stale"

	// The neighbor is known to have been reachable recently (within tens
	// of seconds ago). For IPv6, a positive confirmation was received
	// within the time specified in the ReachabilityTime.ReachableTime
	// member that the forward path to the neighbor was functioning properly.
	// While the State is Reachable, no special action takes place as
	// packets are sent.
	Reachable ARPEntryState = "reachable"

	// The IP address is a permanent address.
	Permanent ARPEntryState = "permanent"

	// specific to linux
	Failed ARPEntryState = "failed"
	NoARP  ARPEntryState = "noarp"
	None   ARPEntryState = "none"
	// Unknown
	Unknown ARPEntryState = "???"
)

func WindowsState(state int) ARPEntryState {
	switch state {
	case 0:
		return Unreachable
	case 1:
		return Incomplete
	case 2:
		return Probe
	case 3:
		return Delay
	case 4:
		return Stale
	case 5:
		return Reachable
	case 6:
		return Permanent
	default:
		return Unknown
	}
}

func LinuxState(state int) ARPEntryState {
	switch state {
	case 0x01:
		return Incomplete
	case 0x02:
		return Reachable
	case 0x04:
		return Stale
	case 0x08:
		return Delay
	case 0x10:
		return Probe
	case 0x20:
		return Failed
	case 0x40:
		return NoARP
	case 0x80:
		return Permanent
	case 0x00:
		return None
	default:
		return Unknown
	}
}
