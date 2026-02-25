//go:build linux

package ping

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// PingSubnet4 pings all provided IPs using a single ICMP socket.
// This avoids the race condition where multiple goroutines with separate
// sockets consume each other's replies.
func PingSubnet4(targets []net.IP, source net.IP, timeout time.Duration, onRecv func(net.IP)) error {
	if len(targets) == 0 {
		return nil
	}

	src := "0.0.0.0"
	if source != nil {
		src = source.String()
	}

	// Try privileged mode first (most reliable)
	protocol := "ip4:icmp"
	conn, err := icmp.ListenPacket("ip4:icmp", src)
	if err != nil {
		// Fall back to unprivileged (less reliable)
		protocol = "udp4"
		conn, err = icmp.ListenPacket(protocol, src)
		if err != nil {
			return fmt.Errorf("cannot create ICMP socket: %w", err)
		}
	}
	defer conn.Close()

	// Track pending targets
	pending := make(map[string]bool)
	for _, ip := range targets {
		if ip[len(ip)-1] == 0xff {
			continue // skip broadcast
		}
		pending[ip.To4().String()] = true
	}

	if len(pending) == 0 {
		return nil
	}

	pid := os.Getpid() & 0xffff

	// Send all pings
	for _, target := range targets {
		if target[len(target)-1] == 0xff {
			continue // skip broadcast
		}

		var dest net.Addr
		if protocol == "ip4:icmp" {
			dest = &net.IPAddr{IP: target}
		} else {
			dest = &net.UDPAddr{IP: target, Port: 0}
		}

		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   pid,
				Seq:  1,
				Data: []byte("situation"),
			},
		}

		msgBytes, err := msg.Marshal(nil)
		if err != nil {
			continue
		}

		_, _ = conn.WriteTo(msgBytes, dest)
	}

	// Set read deadline for the entire receive phase
	deadline := time.Now().Add(timeout)
	if err := conn.SetReadDeadline(deadline); err != nil {
		return err
	}

	// Read replies until timeout
	reply := make([]byte, 1500)
	for {
		n, peer, err := conn.ReadFrom(reply)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break // timeout, we're done
			}
			return err
		}

		// Extract source IP from peer address
		var srcIP net.IP
		switch p := peer.(type) {
		case *net.IPAddr:
			srcIP = p.IP.To4()
		case *net.UDPAddr:
			srcIP = p.IP.To4()
		}

		if srcIP == nil {
			continue
		}

		srcStr := srcIP.String()

		// Check if this is a target we're waiting for
		if !pending[srcStr] {
			continue
		}

		// Parse the ICMP message
		rm, err := icmp.ParseMessage(1, reply[:n])
		if err != nil {
			continue
		}

		// Only process Echo Replies with matching ID/Seq
		if rm.Type == ipv4.ICMPTypeEchoReply {
			if echo, ok := rm.Body.(*icmp.Echo); ok {
				if echo.ID == pid && echo.Seq == 1 {
					delete(pending, srcStr)
					onRecv(srcIP)
				}
			}
		}

		// All targets found
		if len(pending) == 0 {
			break
		}
	}

	return nil
}

// Ping4 pings a single target. For concurrent pinging of multiple targets,
// use PingSubnet4 instead to avoid race conditions with raw sockets.
func Ping4(target net.IP, source net.IP, timeout time.Duration, onRecv func(net.IP)) error {
	return PingSubnet4([]net.IP{target}, source, timeout, onRecv)
}
