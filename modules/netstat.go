package modules

import (
	"fmt"
	"os/user"
	"runtime"

	"github.com/cakturk/go-netstat/netstat"
	"github.com/situation-sh/situation/store"
)

func init() {
	RegisterModule(&NetstatModule{})
}

// Module definition ---------------------------------------------------------

// NetstatModule aims to retrieve infos like the netstat command does
// It must be run as root to retrieve PID/process information. Without
// these data, it is rather hard to build reliable links between open
// ports and programs.
//
// Caveats
// =======
// On windows, the privileges are not checked (because we need to parse
// the SID or another thing maybe). So the module is always run.
type NetstatModule struct{}

func (m *NetstatModule) Name() string {
	return "netstat"
}

func (m *NetstatModule) Dependencies() []string {
	return []string{"host-basic"}
}

func listeningPortFilter(e *netstat.SockTabEntry) bool {
	if e.LocalAddr.IP.IsLoopback() {
		return false
	}
	if e.State != netstat.Listen {
		return false
	}
	return true
}

// helper for the Run function
type netstatProvider func(accept netstat.AcceptFn) ([]netstat.SockTabEntry, error)

func (m *NetstatModule) Run() error {
	logger := GetLogger(m)

	u, err := user.Current()
	if err != nil {
		return err
	}

	if runtime.GOOS == "linux" && u.Uid != "0" {
		return fmt.Errorf("it should be run as root (UID: %s)", u.Uid)
	}

	machine := store.GetHost()
	providers := []netstatProvider{netstat.TCPSocks, netstat.UDPSocks, netstat.TCP6Socks, netstat.UDP6Socks}
	protocols := []string{"tcp", "udp", "tcp6", "udp6"}

	// loop over all providers
	for k, provider := range providers {
		// list all entries by protocol
		if entries, err := provider(listeningPortFilter); err == nil {
			for _, entry := range entries {
				if entry.Process != nil {
					// ignore docker-proxy
					// this process aims at forwarding port
					if entry.Process.Name == "docker-proxy" {
						continue
					}
					soft, created := machine.GetOrCreateApplicationByName(entry.Process.Name)
					if created {
						// logging
						logger.WithField("app", soft.Name).Info("Application found")
					}

					endpoint, created := soft.AddEndpoint(
						entry.LocalAddr.IP,
						entry.LocalAddr.Port,
						protocols[k])
					// fmt.Printf("%+v\n", soft)
					// fmt.Printf("len: %v, cap: %v, %v\n",
					// len(soft.Endpoints), cap(soft.Endpoints), soft.Endpoints)
					if created {
						// fmt.Println(len(soft.Endpoints))
						// logging
						l := logger.WithField("app", soft.Name)
						l = l.WithField("ip", endpoint.Addr)
						l = l.WithField("port", endpoint.Port)
						l = l.WithField("proto", endpoint.Protocol)
						l.Info("Endpoint found")
					}
				}
			}
		}
	}

	return nil
}
