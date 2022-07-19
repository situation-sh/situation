package modules

import (
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/store"
	"github.com/situation-sh/situation/utils"
)

// According to the authors of go-ping
// the pinger must be privileged on windows
// even if we do not run the agent as admin/root
const useICMP = (runtime.GOOS == "windows")

const errorPrefix = "[ERROR_PREFIX]"

func init() {
	RegisterModule(&PingModule{})
}

type PingModule struct{}

func (m *PingModule) Name() string {
	return "ping"
}

func (m *PingModule) Dependencies() []string {
	return []string{"host-network"}
}

func singlePing(ip net.IP, maskSize int, wg *sync.WaitGroup, cerr chan error) {
	defer wg.Done()
	pinger, err := ping.NewPinger(ip.String())
	if err != nil {
		return
	}
	pinger.Count = 1
	pinger.SetPrivileged(useICMP)
	pinger.Timeout = 100 * time.Millisecond

	// callback when a target responds
	pinger.OnRecv = func(p *ping.Packet) {
		logger := logrus.WithField("module", "ping")
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
			logger = logger.WithField("ip", nic.IP).WithField("mask", maskSize)
		} else {
			nic.IP6 = ip.To16()
			nic.Mask6Size = maskSize
			logger = logger.WithField("ip6", nic.IP6).WithField("mask", maskSize)
		}

		// create machine with that NIC
		m = models.NewMachine()
		m.NICS = append(m.NICS, &nic)

		// put this machine to the store
		logger.Info("New machine added")
		store.InsertMachine(m)
	}
	if err := pinger.Run(); err != nil {
		cerr <- fmt.Errorf("error while pinging %v: %v", ip, err)
		return
	}
	pinger.Stop()
}

func pingSubnetwork(network *net.IPNet) error {
	var wg sync.WaitGroup

	cerr := make(chan error)
	for ip := range utils.Iterate(network) {
		// ignore 0 and 255 in case of IPv4
		if utils.IsReserved(ip) {
			continue
		}

		logrus.Debugf("Pinging %s", ip)
		ms, _ := network.Mask.Size()

		wg.Add(1)
		go singlePing(ip, ms, &wg, cerr)
	}

	wg.Wait()

	concatErrors := ""
	for {
		select {
		case err := <-cerr:
			concatErrors += errorPrefix + err.Error()
		default:
			if len(concatErrors) == 0 {
				return nil
			}
			// concat the errors
			return fmt.Errorf(concatErrors)
		}
	}
}

// Ping sends unprivileged ICMP echo messages to all
// hosts on a subnetwork
func (m *PingModule) Run() error {
	logger := GetLogger(m)
	errorMsg := ""

	host := store.GetHost()
	for _, nic := range host.NICS {
		for _, network := range []*net.IPNet{nic.Network()} {
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

			logger.Infof("Pinging %s (%s)", network.String(), nic.Name)
			if err := pingSubnetwork(network); err != nil {
				logger.Error("An error occurred :(")
				errorMsg += fmt.Sprintf("Error(s) occurred while pinging %s:", network.String())
				errorMsg += strings.ReplaceAll(err.Error(), errorPrefix, "\n\t - ")
			}
		}
	}

	if len(errorMsg) > 0 {
		return fmt.Errorf(errorMsg)
	}
	return nil
}
