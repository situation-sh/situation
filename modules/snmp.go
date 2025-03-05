// LINUX(SNMPModule) ok
// WINDOWS(SNMPModule) ok
// MACOS(SNMPModule) ?
// ROOT(SNMPModule) no
package modules

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/situation-sh/situation/modules/snmp"
	"github.com/situation-sh/situation/store"
)

const (
	defaultSNMPVersion   = uint(gosnmp.Version2c)
	defaultSNMPCommunity = "public"
	defaultSNMPTimeout   = 1 * time.Second
	defaultSNMPTransport = "udp"
	defaultSNMPPort      = uint(161)
)

func init() {
	m := &SNMPModule{}
	RegisterModule(m)
	SetDefault(m, "version", defaultSNMPVersion, "SNMP version to use")
	SetDefault(m, "community", defaultSNMPCommunity, "SNMP community to query")
	SetDefault(m, "timeout", defaultSNMPTimeout, "SNMP query timeout")
	SetDefault(m, "transport", defaultSNMPTransport, "TCP or UDP transport protocol")
	SetDefault(m, "port", defaultSNMPPort, "Port to connect")
}

// SNMPModule
// Module to collect data through SNMP protocol.
//
// This module need to access the following OID TREE: `.1.3.6.1.2.1`
// In case of snmpd, the configuration (snmpd.conf) should then include something like this:
//
//	```conf
//	view systemonly included .1.3.6.1.2.1
//	```
type SNMPModule struct{}

func (m *SNMPModule) Name() string {
	return "snmp"
}

func (m *SNMPModule) Dependencies() []string {
	// depends on arp to ensure a rather fresh
	// arp table
	return []string{"arp"}
}

func (m *SNMPModule) Run() error {
	logger := GetLogger(m)
	errs := make([]error, 0)
	// logger := GetLogger(m)
	version, err := GetConfig[uint](m, "version")
	if err != nil {
		return err
	}
	transport, err := GetConfig[string](m, "transport")
	if err != nil {
		return err
	}
	port, err := GetConfig[uint](m, "port")
	if err != nil {
		return err
	}
	timeout, err := GetConfig[time.Duration](m, "timeout")
	if err != nil {
		return err
	}
	community, err := GetConfig[string](m, "community")
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	cerr := make(chan error)
	done := make(chan bool)

	// consume the channel, joins error
	go func() {
		for e := range cerr {
			errs = append(errs, e)
		}
		done <- true
	}()

	for m := range store.IterateMachines() {
		// ignore host machine
		if m.IsHost() {
			continue
		}
		for _, nic := range m.NICS {
			if nic.IP.IsLoopback() || nic.IP.IsMulticast() {
				continue
			}

			g := gosnmp.GoSNMP{
				Target:    nic.IP.String(),
				Version:   gosnmp.SnmpVersion(version),
				Context:   context.Background(),
				Retries:   2,
				Transport: transport,
				Port:      uint16(port),
				Timeout:   timeout,
				Community: community,
			}

			wg.Add(1)
			go snmp.RunSingle(&g, m, &wg, cerr, logger)
		}
	}

	// wait all the snmp request to complete
	wg.Wait()
	// close the error channel (it will end the cosumer above)
	close(cerr)
	// ensure that that the consumer has completed
	<-done

	return errors.Join(errs...)
}
