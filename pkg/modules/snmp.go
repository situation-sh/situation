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

	"github.com/asiffer/puzzle"
	"github.com/gosnmp/gosnmp"
	"github.com/situation-sh/situation/pkg/modules/snmp"
)

// const (
// 	defaultSNMPVersion   = uint(gosnmp.Version2c)
// 	defaultSNMPCommunity = "public"
// 	defaultSNMPTimeout   = 1 * time.Second
// 	defaultSNMPTransport = "udp"
// 	defaultSNMPPort      = uint(161)
// )

func init() {
	m := &SNMPModule{
		Version:   uint8(gosnmp.Version2c),
		Community: "public",
		Timeout:   1 * time.Second,
		Transport: "udp",
		Port:      161,
	}
	registerModule(m)
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
type SNMPModule struct {
	BaseModule
	Version   uint8
	Community string
	Timeout   time.Duration
	Transport string
	Port      uint16
}

func (m *SNMPModule) Bind(config *puzzle.Config) error {
	if err := setDefault(config, m, "version", &m.Version, "SNMP version to use"); err != nil {
		return err
	}
	if err := setDefault(config, m, "community", &m.Community, "SNMP community to query"); err != nil {
		return err
	}
	if err := setDefault(config, m, "timeout", &m.Timeout, "SNMP query timeout"); err != nil {
		return err
	}
	if err := setDefault(config, m, "transport", &m.Transport, "TCP or UDP transport protocol"); err != nil {
		return err
	}
	if err := setDefault(config, m, "port", &m.Port, "Port to connect"); err != nil {
		return err
	}
	return nil
}

func (m *SNMPModule) Name() string {
	return "snmp"
}

func (m *SNMPModule) Dependencies() []string {
	// depends on arp to ensure a rather fresh
	// arp table
	return []string{"arp"}
}

func (m *SNMPModule) Run() error {

	errs := make([]error, 0)

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

	for machine := range m.store.IterateMachines() {
		// ignore host machine
		if m.store.IsHost(machine) {
			continue
		}
		for _, nic := range machine.NICS {
			if nic.IP.IsLoopback() || nic.IP.IsMulticast() {
				continue
			}

			g := gosnmp.GoSNMP{
				Target:    nic.IP.String(),
				Version:   gosnmp.SnmpVersion(m.Version),
				Context:   context.Background(),
				Retries:   2,
				Transport: m.Transport,
				Port:      m.Port,
				Timeout:   m.Timeout,
				Community: m.Community,
			}

			wg.Add(1)
			go snmp.RunSingle(&g, machine, &wg, cerr, m.logger, m.store)
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
