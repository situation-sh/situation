package utils

import (
	"net"
)

// CopyIP returns a new buffer containing
// the input IP address
func CopyIP(ip net.IP) net.IP {
	tmp := make([]byte, len(ip))
	copy(tmp, ip)
	return tmp
}

// BaseNetwork returns the strict CIDR network
// Example 192.168.1.15/24 -> 192.168.1.0/24
// Seems useless
// func BaseNetwork(n *net.IPNet) *net.IPNet {
// 	return &net.IPNet{
// 		IP:   n.IP.Mask(n.Mask),
// 		Mask: n.Mask,
// 	}
// }

// Iterate returns a channel yielding IP addresses
// included in IP network. It makes copies to avoid
// concurrency issues while reading the received values
func Iterate(n *net.IPNet) chan net.IP {
	// create a copy of the IP address setting
	// free bits to zero
	base := n.IP.Mask(n.Mask)
	// init channels
	c := make(chan net.IP)
	//
	size := len(base)
	// get the mask
	frozen, total := n.Mask.Size()
	// number of IP (2^(n-k))
	nIP := 1 << uint(total-frozen)
	// run
	go func() {
		c <- CopyIP(base)
		for i := 0; i < nIP-1; i++ {
			// byte index (starting from the end)
			k := 1
			// increment last byte
			base[size-k]++
			for base[size-k] == 0 {
				// increment last byte if 255 is reached
				k++
				base[size-k]++
			}
			// send buffer
			c <- CopyIP(base)
		}
		// close
		close(c)
	}()

	return c
}

// ExtractNetworks returns the address (sub-networks actually)
// attached to an interface. If keepLocalAddress is set to True
// the IP field of net.IPNet instance is set to the local IP address
func ExtractNetworks(iface *net.Interface, keepLocalAddress bool) []*net.IPNet {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil
	}

	nets := make([]*net.IPNet, 0)
	for _, addr := range addrs {
		ip, ipnet, err := net.ParseCIDR(addr.String())
		if err != nil {
			continue
		}
		if keepLocalAddress {
			ipnet.IP = ip
		}
		nets = append(nets, ipnet)
	}

	return nets
}

// PreferredNetwork returns the (non-strict) network
// where your system sends outgoing request.
// Seems useless
// func PreferredNetwork() (*net.IPNet, error) {
// 	// call google
// 	ip := net.ParseIP("8.8.8.8")

// 	router, err := routing.New()
// 	if err != nil {
// 		return nil, errors.Wrap(err, "error while creating routing object")
// 	}

// 	iface, _, _, err := router.Route(ip)
// 	if err != nil {
// 		return nil, errors.Wrapf(err, "error routing to ip: %s", ip)
// 	}

// 	networks := ExtractNetworks(iface, true)
// 	return networks[0], nil
// }

// Seems useless
// func Uint32ToIP(u uint32) net.IP {
// 	ip := make(net.IP, 4)
// 	binary.LittleEndian.PutUint32(ip, u)
// 	return ip
// }

func IsReserved(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[3] == 0 || ip4[3] == 255
	}
	return false
}

func EnforceMask(nw *net.IPNet) {
	if ip4 := nw.IP.To4(); ip4 != nil {
		for i := 0; i < 4; i++ {
			ip4[i] &= nw.Mask[i]
		}
		nw.IP = ip4
	} else {
		for i := 0; i < 16; i++ {
			nw.IP[i] &= nw.Mask[i]
		}
	}
}
