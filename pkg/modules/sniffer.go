//go:build linux && cgo
// +build linux,cgo

// LINUX(SnifferModule) ok
// WINDOWS(SnifferModule) ?
// MACOS(SnifferModule) ?
// ROOT(SnifferModule) yes
package modules

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"net"
	"sync"
	"time"

	"github.com/asiffer/puzzle"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	_ "github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/store"
	"github.com/situation-sh/situation/pkg/utils"
)

func init() {
	registerModule(&SnifferModule{
		Ifaces:      []string{},
		Promisc:     true,
		BPFFilter:   "",
		Duration:    5 * time.Second,
		snapshotLen: 1024,
	})
}

type SnifferModule struct {
	BaseModule

	Ifaces      []string
	Promisc     bool
	BPFFilter   string
	Duration    time.Duration
	snapshotLen int32 // not exposed to user
}

func (m *SnifferModule) Bind(config *puzzle.Config) error {
	if err := setDefault(config, m, "ifaces", &m.Ifaces, "Filter network interfaces to sniff (sniff all by default)"); err != nil {
		return err
	}
	if err := setDefault(config, m, "promisc", &m.Promisc, "Enable promiscuous mode (need root or capabilities)"); err != nil {
		return err
	}
	if err := setDefault(config, m, "bpf", &m.BPFFilter, "Add optional BPF filter"); err != nil {
		return err
	}
	if err := setDefault(config, m, "duration", &m.Duration, "Sniffing duration"); err != nil {
		return err
	}
	return nil
}

func (m *SnifferModule) Name() string {
	return "sniffer"
}

func (m *SnifferModule) Dependencies() []string {
	return []string{"tls"}
}

func (m *SnifferModule) Run() error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	//
	filtered := make([]net.Interface, 0)
	if len(m.Ifaces) > 0 {
		filtered := []net.Interface{}
		for _, iface := range ifaces {
			for _, name := range m.Ifaces {
				if iface.Name == name {
					filtered = append(filtered, iface)
				}
			}
		}
	} else {
		for _, iface := range ifaces {
			if (iface.Flags & net.FlagUp) == 0 {
				continue
			}
			if (iface.Flags & net.FlagLoopback) != 0 {
				continue
			}
			if (iface.Flags & net.FlagRunning) == 0 {
				continue
			}
			// fmt.Println(iface)
			filtered = append(filtered, iface)
		}
	}

	if len(filtered) == 0 {
		m.logger.Warn("No network interface to sniff")
		return fmt.Errorf("No network interface to sniff")
	}

	m.logger.Errorf("Sniffing on interfaces: %v", filtered)
	return m.ConcurrentSniff(filtered)
}

func handleFlow(flow models.Flow, store store.Store) {
	// find the machine that has the local address (the caller)
	if !utils.IsPublic(flow.LocalAddr) {
		machine := store.GetMachineByIP(flow.LocalAddr)
		// if the machine already exists
		if machine == nil {
			machine = models.NewMachine()
			nic := models.NetworkInterface{}
			if flow.LocalAddr.To4() != nil {
				nic.IP = flow.LocalAddr.To4()
			} else {
				nic.IP6 = flow.LocalAddr.To16()
			}
			machine.NICS = append(machine.NICS, &nic)

			pkg := models.NewPackage()
			app := models.NewApplication()
			app.Flows = append(app.Flows, &flow)
			pkg.Applications = append(pkg.Applications, app)
			machine.InsertPackage(pkg)

			store.InsertMachine(machine)
		}
	}

	// we get-or-create a remote machine if it is not a public service
	if !utils.IsPublic(flow.RemoteAddr) {
		remote := store.GetMachineByIP(flow.RemoteAddr)
		if remote == nil {
			remote = models.NewMachine()
			remote.NICS = append(remote.NICS, &models.NetworkInterface{
				IP: flow.RemoteAddr,
			})
			store.InsertMachine(remote)
		}
		app, _ := remote.GetOrCreateApplicationByEndpoint(
			flow.RemotePort,
			flow.Protocol,
			flow.RemoteAddr,
		)
		// we add the flow to the remote endpoint but reverted
		rf := flow.Revert()
		app.AddFlow(&rf)
	}

}

func (m *SnifferModule) singleSniff(iface net.Interface, wg *sync.WaitGroup, errChan chan error, flowChan chan models.Flow) {
	defer wg.Done()
	flows, err := m.Sniff(iface)
	// fmt.Println("Sniff returns")
	if err != nil {
		errChan <- err
		return
	}
	// fmt.Println("sending flows")
	for _, flow := range flows {
		flowChan <- flow
	}

}

func (m *SnifferModule) ConcurrentSniff(ifaces []net.Interface) error {
	errChan := make(chan error)
	flowChan := make(chan models.Flow)
	var wg sync.WaitGroup

	errs := make([]error, 0)

	// start a goroutine to collect errors
	go func() {
		for err := range errChan {
			errs = append(errs, err)
		}
	}()

	// start a goroutine to collect flows
	go func() {
		for flow := range flowChan {
			// process the flow
			handleFlow(flow, m.store)
		}
	}()

	// start sniffing on each interface concurrently
	for _, iface := range ifaces {
		wg.Add(1)
		go m.singleSniff(iface, &wg, errChan, flowChan)
	}

	m.logger.Error("Waiting for all sniffing routines to finish...")
	wg.Wait()
	close(errChan)
	close(flowChan)
	return errors.Join(errs...)
}

func hashFlow(network *gopacket.Flow, transport *gopacket.Flow) uint64 {
	var buf [16]byte
	nh := network.FastHash()
	th := transport.FastHash()
	h := fnv.New64a()
	binary.LittleEndian.PutUint64(buf[0:], nh)
	binary.LittleEndian.PutUint64(buf[8:], th)
	h.Write(buf[:])
	return h.Sum64()
}

func (m *SnifferModule) Sniff(iface net.Interface) (map[uint64]models.Flow, error) {
	m.logger.Debugf("Opening network interface: %s", iface.Name)

	handle, err := pcap.OpenLive(iface.Name, m.snapshotLen, m.Promisc, pcap.BlockForever)
	if err != nil {
		return nil, err
	}
	defer handle.Close()

	// set BPF filter if any
	if len(m.BPFFilter) > 0 {
		m.logger.Debugf("Setting BPF filter: %s", m.BPFFilter)
		err = handle.SetBPFFilter(m.BPFFilter)
		if err != nil {
			return nil, fmt.Errorf("error setting BPF filter: %v", err)
		}
	}
	// handle, err := afpacket.NewTPacket(
	// 	afpacket.OptInterface(iface.Name),
	// 	afpacket.OptFrameSize(m.snapshotLen),
	// )
	// if err != nil {
	// 	return nil, fmt.Errorf("error opening afpacket on interface %s: %v", iface.Name, err)
	// }
	// defer handle.Close()
	// set BPF filter if any
	// if len(m.BPFFilter) > 0 {
	// 	// af.
	// 	pcap.CompileBPFFilter(handle.)
	// 	m.logger.Debugf("Setting BPF filter: %s", m.BPFFilter)
	// 	// err = af.SetBPF()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error setting BPF filter: %v", err)
	// 	}
	// }

	// Channel to stop capture after timeout
	stop := time.After(m.Duration)

	var eth layers.Ethernet
	var ip4 layers.IPv4
	var ip6 layers.IPv6
	var tcp layers.TCP
	var udp layers.UDP

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp, &udp)
	decodedLayers := make([]gopacket.LayerType, 0, 10)

	macAddresses := make(map[string]bool)
	flows := make(map[uint64]models.Flow)
	// structure := packetStructure{}

	for {
		select {
		case <-stop:
			m.logger.Warnf("Stopping sniffing on interface: %s", iface.Name)
			return flows, nil
		default:
			data, _, err := handle.ZeroCopyReadPacketData()
			if err != nil {
				m.logger.Warnf("Fail to read packet: %v", err)
				continue
			}

			if err := parser.DecodeLayers(data, &decodedLayers); err != nil {
				m.logger.Debugf("Failed to decode packet layer: %v", err)
			}

			// init a real flow
			flow := models.Flow{Status: "ESTABLISHED"}
			nf := gopacket.Flow{}
			tf := gopacket.Flow{}
			revert := false

			// analyze layers
			for _, layerType := range decodedLayers {
				switch layerType {
				case layers.LayerTypeEthernet:
					macAddresses[eth.SrcMAC.String()] = true
					macAddresses[eth.DstMAC.String()] = true
				case layers.LayerTypeIPv4:
					flow.LocalAddr = ip4.SrcIP.To4()
					flow.RemoteAddr = ip4.DstIP.To4()
					nf = ip4.NetworkFlow()
				case layers.LayerTypeIPv6:
					flow.LocalAddr = ip6.SrcIP.To16()
					flow.RemoteAddr = ip6.DstIP.To16()
					nf = ip6.NetworkFlow()
				case layers.LayerTypeTCP:
					flow.Protocol = "tcp"
					flow.LocalPort = uint16(tcp.SrcPort)
					flow.RemotePort = uint16(tcp.DstPort)
					tf = tcp.TransportFlow()
					revert = tcp.SYN && tcp.ACK // it means that the server is answering
				case layers.LayerTypeUDP:
					flow.Protocol = "udp"
					flow.LocalPort = uint16(udp.SrcPort)
					flow.RemotePort = uint16(udp.DstPort)
					tf = udp.TransportFlow()
					// try to detect direction based on port number
					// https://en.wikipedia.org/wiki/Registered_port
					if flow.LocalPort >= 49152 {
						revert = true
					}
				}
			}

			// handle flow only if we have a known transport protocol
			if flow.Protocol != "" {
				h := hashFlow(&nf, &tf)
				if _, exists := flows[h]; !exists {
					m.logger.WithField("flow-local-addr", flow.LocalAddr).
						WithField("flow-local-port", flow.LocalPort).
						WithField("flow-remote-addr", flow.RemoteAddr).
						WithField("flow-remote-port", flow.RemotePort).
						WithField("flow-proto", flow.Protocol).
						WithField("flow-status", flow.Status).
						Info("Flow found")
					// we try to ensure that LocalAddr is the caller and RemoteAddr the callee
					if revert {
						flows[h] = flow.Revert()
					} else {
						flows[h] = flow
					}
				}
			}

		}
	}

	// source := gopacket.NewPacketSource(handle, handle.LinkType())
	// packets := source.Packets()

	// for {
	// 	select {
	// 	case packet := <-packets:
	// 		n := packet.NetworkLayer()
	// 		n.
	// 		flow := packet.NetworkLayer().NetworkFlow()
	// 		handleFlow(flow)
	// 	case <-stop:
	// 		return nil
	// 	}
	// }

	// var eth layers.Ethernet
	// var ip4 layers.IPv4
	// var ip6 layers.IPv6
	// var tcp layers.TCP
	// var udp layers.UDP

	// parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp, &udp)
	// decodedLayers := make([]gopacket.LayerType, 0, 10)

	// source := gopacket.NewPacketSource(handle, handle.LinkType())
	// source.NoCopy = true

	// for {
	// 	data, _, err := handle.ReadPacketData()
	// 	if err != nil {
	// 		m.logger.Warnf("Fail to read packet: %v", err)
	// 		continue
	// 	}
	// 	if err := parser.DecodeLayers(data, &decodedLayers); err != nil {
	// 		m.logger.Warnf("Failed to decode packet: %v", err)
	// 		continue
	// 	}
	// 	for _, layerType := range decoded {
	// 		switch layerType {
	// 		case layers.LayerTypeIPv6:
	// 			fmt.Println("    IP6 ", ip6.SrcIP, ip6.DstIP)
	// 		case layers.LayerTypeIPv4:
	// 			fmt.Println("    IP4 ", ip4.SrcIP, ip4.DstIP)
	// 		}
	// 	}
	// }

}
