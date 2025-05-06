// LINUX(PingModule) ok
// WINDOWS(PingModule) ok
// MACOS(PingModule) ?
// ROOT(PingModule) no
package modules

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	ping "github.com/prometheus-community/pro-bing"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/modules/pingconfig"
	"github.com/situation-sh/situation/store"
	"github.com/situation-sh/situation/utils"
)

const errorPrefix = "[ERROR_PREFIX]"

func init() {
	RegisterModule(&PingModule{})
}

// PingModule pings local networks to discover new hosts.
//
// The module relies on [pro-bing] library.
//
// A single ping attempt is made on every host of the local networks
// (the host may belong to several networks). Only IPv4 networks with
// prefix length >=20 are treated.
// The ping timeout is hardset to 300ms.
//
// [pro-bing]: https://github.com/prometheus-community/pro-bing
type PingModule struct{}

func (m *PingModule) Name() string {
	return "ping"
}

func (m *PingModule) Dependencies() []string {
	return []string{"host-network"}
}

func singlePing(ip net.IP, maskSize int, wg *sync.WaitGroup, cerr chan<- error, log *logrus.Entry) {
	defer wg.Done()
	privileged := pingconfig.UseICMP()
	pinger, err := ping.NewPinger(ip.String())
	if err != nil {
		return
	}
	pinger.Count = 1
	pinger.SetPrivileged(privileged)
	pinger.Timeout = 300 * time.Millisecond
	// see https://github.com/go-ping/ping/issues/168
	pinger.Size = 2048

	// callback when a target responds
	pinger.OnRecv = func(p *ping.Packet) {
		ip := p.IPAddr.IP

		// check the store
		m := store.GetMachineByIP(ip)
		if m != nil {
			return
		}

		// create nic
		nic := models.NetworkInterface{}
		if ip4 := ip.To4(); ip4 != nil {
			nic.IP = ip4
			nic.MaskSize = maskSize
			log = log.WithField("ip", nic.IP).WithField("mask", maskSize)
		} else {
			nic.IP6 = ip.To16()
			nic.Mask6Size = maskSize
			log = log.WithField("ip6", nic.IP6).WithField("mask", maskSize)
		}

		// create machine with that NIC
		m = models.NewMachine()
		m.NICS = append(m.NICS, &nic)

		// put this machine to the store
		log.Info("New machine added")
		store.InsertMachine(m)
	}

	err = pinger.Run()
	if err == nil {
		pinger.Stop()
		return
	}

	log.Debugf("ping with privileged=%v failed (%v), trying with the opposite", privileged, err)
	pinger.SetPrivileged(!privileged)

	err = pinger.Run()
	defer pinger.Stop()

	if err == nil {
		return
	}
	log.Errorf("ping with privileged=%v failed (%v), no fallback", privileged, err)
	cerr <- fmt.Errorf("error while pinging %v: %v", ip, err)
}

func pingSubnetwork(network *net.IPNet, log *logrus.Entry) error {
	var wg sync.WaitGroup
	cerr := make(chan error)
	done := make(chan bool)
	var err error
	var once sync.Once

	// consume the channel, sets the first error only
	go func() {
		for e := range cerr {
			once.Do(func() { err = e })
		}
		done <- true
	}()

	for ip := range utils.Iterate(network) {
		// ignore 0 and 255 in case of IPv4
		if utils.IsReserved(ip) {
			continue
		}

		log.Debugf("Pinging %s", ip)
		ms, _ := network.Mask.Size()

		wg.Add(1)
		go singlePing(ip, ms, &wg, cerr, log)
	}

	log.Debug("Waiting ping to finish")
	wg.Wait()

	// closing the channel ensures that the above goroutine
	// will stop. Indeed, if no error raised, err will be equal
	// to nil
	close(cerr)
	<-done

	// now the goroutine is over
	// err is well sync
	return err
}

// Ping sends unprivileged ICMP echo messages to all
// hosts on a subnetwork
func (m *PingModule) Run() error {
	logger := GetLogger(m)
	errorMsg := ""

	// host := store.GetHost()
	// try to ping all networks
	for _, network := range store.GetAllIPv4Networks() {
		// network() returns the IPv4 network attached to this nic
		// for _, network := range []*net.IPNet{nic.Network()} {
		if network == nil {
			continue
		}

		switch ones, bits := network.Mask.Size(); {
		case ones < 20:
			// ignore to large network (here /20 at most)
			logger.Warnf("Ignoring %s (network is too wide)", network.String())
			continue
		case ones > 24:
			// if the network is restricted. We try to
			// send pings in a wider one. It may appear
			// in VPN cases
			// this change does not modify the mask inside
			// the store
			network.Mask = net.CIDRMask(24, bits)
		}

		logger.Infof("Pinging %s", network.String())
		if err := pingSubnetwork(network, logger); err != nil {
			logger.Error(err)
			errorMsg += fmt.Sprintf("Error(s) occurred while pinging %s:", network.String())
			errorMsg += strings.ReplaceAll(err.Error(), errorPrefix, "\n\t - ")
		}
		// }
	}

	if len(errorMsg) > 0 {
		return fmt.Errorf("%s", errorMsg)
	}
	return nil
}
