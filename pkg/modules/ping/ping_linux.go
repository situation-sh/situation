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

func Ping4(target net.IP, source net.IP, timeout time.Duration, onRecv func(net.IP)) error {
	if target[3] == 0xff {
		// skip broadcast addresses
		return nil
	}
	src := "0.0.0.0"
	if source != nil {
		src = source.String()
	}
	// Use "udp4" instead of "ip4:icmp" for unprivileged ping
	// Try privileged mode first (most reliable)
	protocol := "ip4:icmp"
	var dest net.Addr = &net.IPAddr{IP: target}
	conn, err := icmp.ListenPacket("ip4:icmp", src)
	if err != nil {
		// Fall back to unprivileged (less reliable)
		protocol = "udp4"
		dest = &net.UDPAddr{IP: target}
		conn, err = icmp.ListenPacket(protocol, src)
		if err != nil {
			return fmt.Errorf("cannot create ICMP socket: %w", err)
		}
	}
	defer conn.Close()

	// Create ICMP Echo Request
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("situation"),
		},
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return err
	}

	// send
	_, err = conn.WriteTo(msgBytes, dest)
	if err != nil {
		return err
	}

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(timeout))

	// Wait for reply
	reply := make([]byte, 1500)
	_, _, err = conn.ReadFrom(reply)
	// duration := time.Since(start)

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil
			// return fmt.Errorf("timeout error (%v)", err)
		}
		return err
	}

	// Parse reply
	onRecv(target)

	return nil
}
