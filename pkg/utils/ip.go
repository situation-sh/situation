package utils

import (
	"errors"
	"net"
	"sync"
)

// CopyIP returns a new buffer containing
// the input IP address
func CopyIP(ip net.IP) net.IP {
	tmp := make([]byte, len(ip))
	copy(tmp, ip)
	return tmp
}

func MaskSize(n *net.IPNet) int {
	ones, _ := n.Mask.Size()
	return ones
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
	nIP := 1 << (total - frozen)
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

func ListIPs(n *net.IPNet) []net.IP {
	ips := make([]net.IP, 0)
	for ip := range Iterate(n) {
		ips = append(ips, ip)
	}
	return ips
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

func EnforceMask(nw *net.IPNet) *net.IPNet {
	out := net.IPNet{
		IP:   nw.IP.Mask(nw.Mask),
		Mask: nw.Mask,
	}
	return &out
}

func IsPublic(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if net.IP.Equal(ip, net.IPv4zero) || net.IP.Equal(ip, net.IPv6zero) {
		return false
	}
	return !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() && !ip.IsMulticast() && !ip.IsInterfaceLocalMulticast()
}

type IPWorkerPool struct {
	poolSize uint
	worker   func(ip <-chan net.IP, errChan chan<- error, wg *sync.WaitGroup)
}

func NewIPWorkerPool(poolSize uint, fun func(net.IP) error) *IPWorkerPool {
	worker := func(ipChan <-chan net.IP, errChan chan<- error, wg *sync.WaitGroup) {
		defer wg.Done()
		for i := range ipChan {
			if err := fun(i); err != nil {
				errChan <- err
			}
		}
	}
	return &IPWorkerPool{
		poolSize: poolSize,
		worker:   worker,
	}
}

func (wp *IPWorkerPool) Run(ips []net.IP) error {
	var wg sync.WaitGroup

	ipChan := make(chan net.IP, 2*wp.poolSize)
	errChan := make(chan error)
	errResult := make(chan error)

	// Error consumer
	go func() {
		var errs error
		for err := range errChan {
			errs = errors.Join(errs, err)
		}
		errResult <- errs
	}()
	// for errChan, we need a buffered channel to avoid deadlocks
	// since errors are consumed in the end
	// errChan := make(chan error, len(ips))

	// start workers
	for i := uint(0); i < wp.poolSize; i++ {
		wg.Add(1)
		go wp.worker(ipChan, errChan, &wg)
	}
	// send jobs
	for _, ip := range ips {
		ipChan <- ip
	}
	close(ipChan)
	// wait for workers to finish
	wg.Wait()
	// close errors
	close(errChan)
	return <-errResult
}

func IPVersionFromCIDR(cidr string) int {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0
	}
	if ip.To4() != nil {
		return 4
	}
	return 6
}
