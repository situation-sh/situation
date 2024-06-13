// LINUX(NetstatModule) ok
// WINDOWS(NetstatModule) ok
// MACOS(NetstatModule) ?
// ROOT(NetstatModule) yes
package modules

import (
	"os/user"
	"runtime"

	"github.com/cakturk/go-netstat/netstat"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/store"
	"github.com/situation-sh/situation/utils"
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
// This module is then able to create flows between applications according
// to the tuple (src, srcport, dst, dstport).
//
// On windows, the privileges are not checked (because we need to parse
// the SID or another thing maybe). So the module is always run.
type NetstatModule struct{}

func (m *NetstatModule) Name() string {
	return "netstat"
}

func (m *NetstatModule) Dependencies() []string {
	return []string{"host-basic"}
}

func flowFilter(state netstat.SkState) bool {
	for _, s := range []netstat.SkState{
		netstat.Established,
		netstat.FinWait1,
		netstat.FinWait2,
		netstat.TimeWait,
		netstat.CloseWait,
		netstat.LastAck,
		netstat.Closing} {
		if s == state {
			return true
		}
	}
	return false
}

// portFilter returns true when the connection is listening, established or close-wait
func portFilter(e *netstat.SockTabEntry) bool {
	if e.LocalAddr.IP.IsLoopback() && !e.RemoteAddr.IP.IsLoopback() {
		return false
	}
	if e.State == netstat.Listen || flowFilter(e.State) {
		return true
	}
	return false
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
		logger.Warnf("On Linux, the %s module must be run as root", m.Name())
		return &mustBeRunAsRootError{uid: u.Uid}
	}

	machine := store.GetHost()
	providers := []netstatProvider{netstat.TCPSocks, netstat.UDPSocks, netstat.TCP6Socks, netstat.UDP6Socks}
	protocols := []string{"tcp", "udp", "tcp6", "udp6"}

	// loop over all providers
	for k, provider := range providers {
		// list all entries by protocol
		if entries, err := provider(portFilter); err == nil {
			for _, entry := range entries {
				if entry.Process != nil {
					// ignore docker-proxy
					// this process aims at forwarding port
					if entry.Process.Name == "docker-proxy" {
						continue
					}

					// (NEW!) create localhost if needed (localhost communication)
					if entry.LocalAddr.IP.IsLoopback() && entry.RemoteAddr.IP.IsLoopback() {
						machine.GetOrCreateHostLoopback(entry.LocalAddr.IP)
					}

					name := entry.Process.Name
					args, err := utils.GetCmd(entry.Process.Pid)
					if err == nil && len(args) > 0 {
						name = args[0]
						args = args[1:]
					}
					soft, created := machine.GetOrCreateApplicationByName(name)
					soft.PID = uint(entry.Process.Pid)

					if created {
						// logging
						logger.WithField("app", soft.Name).
							WithField("pid", soft.PID).
							Info("Application found")
					}

					// NEW: add flows
					if flowFilter(entry.State) {
						flow := models.Flow{
							LocalAddr:  entry.LocalAddr.IP,
							RemoteAddr: entry.RemoteAddr.IP,
							LocalPort:  entry.LocalAddr.Port,
							RemotePort: entry.RemoteAddr.Port,
							Protocol:   protocols[k],
							Status:     entry.State.String(),
						}
						soft.Flows = append(soft.Flows, &flow)
						logger.WithField("app", soft.Name).
							WithField("flow-local-addr", flow.LocalAddr).
							WithField("flow-local-port", flow.LocalPort).
							WithField("flow-remote-addr", flow.RemoteAddr).
							WithField("flow-remote-port", flow.RemotePort).
							WithField("flow-proto", flow.Protocol).
							WithField("flow-status", flow.Status).
							Info("Flow found")
					}

					// add args
					if len(args) > 0 {
						soft.Args = args
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
