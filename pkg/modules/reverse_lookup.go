// LINUX(ReverseLookupModule) ok
// WINDOWS(ReverseLookupModule) ok
// MACOS(ReverseLookupModule) ?
// ROOT(ReverseLookupModule) ?
package modules

import (
	"net"
	"strings"
)

func init() {
	registerModule(&ReverseLookupModule{})
}

// ReverseLookupModule tries to get a hostname attached to a local IP address
type ReverseLookupModule struct {
	BaseModule
}

func (m *ReverseLookupModule) Name() string {
	return "reverse-lookup"
}

func (m *ReverseLookupModule) Dependencies() []string {
	// depends on ping to ensure a rather fresh
	// arp table
	return []string{"netstat"}
}

func (m *ReverseLookupModule) Run() error {
	

outer:
	for machine := range m.store.IterateMachines() {
		if machine.Hostname == "" {
			for _, nic := range machine.NICS {
				for _, ip := range []net.IP{nic.IP, nic.IP6} {
					if ip != nil && ip.IsPrivate() {
						// run first lookup
						net.LookupAddr(nic.IP.String()) // #nosec G104 -- we don't care about the errors here
						names, err := net.LookupAddr(nic.IP.String())

						if err != nil {
							m.logger.Error(err)
							continue
						}
						if len(names) > 0 {
							// on linux we may have an ending dot
							machine.Hostname = strings.TrimSuffix(names[0], ".")
							m.logger.WithField("hostname", machine.Hostname).
								WithField("ip", ip).Infof("hostname resolved")
							// go to the next machine
							continue outer
						}
					}
				}
			}
		}

	}
	return nil
}
