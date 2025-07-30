// LINUX(VMwareModule) ok
// WINDOWS(VMwareModule) ok
// MACOS(VMwareModule) ?
// ROOT(VMwareModule) no
package modules

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/store"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func init() {
	m := &VMwareModule{
		Username: "root",
	}
	RegisterModule(m)
	SetDefault(m, "username", &m.Username, "VMware API username")
	SetDefault(m, "password", &m.Password, "VMware API password")
}

// Module definition ---------------------------------------------------------

// VMwareModule tries to connect to esxi/vcenter hosts and list VMs
type VMwareModule struct {
	Username string
	Password string
}

func (m *VMwareModule) Name() string {
	return "vmware"
}

func (m *VMwareModule) Dependencies() []string {
	return []string{"tls"}
}

func (m *VMwareModule) Run() error {
	logger := GetLogger(m)
	var moVM mo.VirtualMachine

	for machine := range store.IterateMachines() {
		for _, app := range machine.Applications() {
			for _, endpoint := range app.Endpoints {
				if endpoint.TLS == nil {
					continue
				}

				target := net.JoinHostPort(endpoint.Addr.String(), fmt.Sprintf("%d", endpoint.Port))
				u, err := url.Parse(fmt.Sprintf("https://%s/sdk", target))
				if err != nil {
					logger.WithError(err).
						WithField("ip", endpoint.Addr).
						WithField("port", endpoint.Port).
						Error("fail to parse URL")
					continue
				}
				u.User = url.UserPassword(m.Username, m.Password)

				ctx := context.Background()
				c, err := govmomi.NewClient(ctx, u, true)
				if err != nil {
					logger.WithError(err).
						WithField("ip", endpoint.Addr).
						WithField("port", endpoint.Port).
						Error("failed to connect")
					continue
				}

				// Finder to search objects
				finder := find.NewFinder(c.Client, true)

				// Set datacenter
				dc, err := finder.DefaultDatacenter(ctx)
				if err != nil {
					logger.WithError(err).
						WithField("ip", endpoint.Addr).
						WithField("port", endpoint.Port).
						Error("failed to get datacenter")
					continue
				}
				finder.SetDatacenter(dc)

				// List all virtual machines
				vms, err := finder.VirtualMachineList(ctx, "*")
				if err != nil {
					logger.WithError(err).
						WithField("ip", endpoint.Addr).
						WithField("port", endpoint.Port).
						Error("failed to list VMs")
					continue
				} else {
					logger.WithField("ip", endpoint.Addr).
						WithField("port", endpoint.Port).
						WithField("vm", len(vms)).
						Infof("Found virtual machines")
				}

				for _, vm := range vms {
					err := vm.Properties(
						ctx,
						vm.Reference(),
						[]string{"name", "summary", "guest", "storage", "runtime", "config"},
						&moVM,
					)
					if err != nil {
						logger.WithError(err).
							WithField("ip", endpoint.Addr).
							WithField("port", endpoint.Port).
							WithField("vm", vm.Name()).
							Error("failed to get VM properties")
						continue
					}
					updateMachine(machine, &moVM, logger.WithField("vm", vm.Name()))
				}

			}
		}
	}
	return nil
}

func updateMachine(parent *models.Machine, vm *mo.VirtualMachine, logger *logrus.Entry) {
	var err error
	if vm == nil {
		return
	}
	if vm.Summary.Runtime.PowerState != types.VirtualMachinePowerStatePoweredOn {
		logger.Debug("VM is not powered on, skipping")
		return
	}
	guest := vm.Guest
	if guest == nil {
		logger.Warn("VM guest information is nil")
		return
	}
	networks := guest.Net
	if len(networks) == 0 {
		logger.Warn("VM guest network information is empty")
		return
	}

	// find if the VM is already in the store
	var machine *models.Machine

netloop:
	for _, n := range networks {
		var mac net.HardwareAddr
		if n.MacAddress != "" {
			mac, err = net.ParseMAC(n.MacAddress)
			if err != nil {
				logger.WithError(err).
					WithField("mac", n.MacAddress).
					Error("failed to parse MAC address")
				mac = nil
			}
		}
		for _, ipstr := range n.IpAddress {
			ip := net.ParseIP(ipstr)
			if ip == nil && mac == nil {
				continue // Skip if both IP and MAC are invalid
			}
			machine = store.GetMachineByNetwork(ip, mac)
			if machine != nil {
				break netloop
			}

		}

	}
	if machine == nil {
		logger.Debug("VM not found in store, creating new machine")
		machine = models.NewMachine()
	}
	// here we do have a machine

	machine.Chassis = "vm"
	machine.ParentMachine = parent.InternalID
	if machine.Arch == "" && strings.Contains(guest.GuestFullName, "64-bit") {
		machine.Arch = "x86_64"
	}
	if machine.Hostname == "" {
		machine.Hostname = guest.HostName
	}
	machine.Uptime = time.Duration(vm.Summary.QuickStats.UptimeSeconds)
	machine.Platform = strings.Replace(guest.GuestFamily, "Guest", "", 1) //gives "windows", "linux"
	re := regexp.MustCompile(`\s*\(.*\)`)
	machine.Distribution = strings.TrimSpace(re.ReplaceAllString(guest.GuestFullName, ""))

	logger.WithField("hostname", machine.Hostname).
		WithField("platform", machine.Platform).
		WithField("distribution", machine.Distribution).
		WithField("arch", machine.Arch).
		Info("VM detected")

	if types.VirtualMachineToolsRunningStatus(guest.ToolsRunningStatus) == types.VirtualMachineToolsRunningStatusGuestToolsRunning {
		// do things only if the tools are running
		// for instance we can run a program to grab the DistributionVersion
	}

	config := vm.Config
	if config != nil {
		// CPU infos
		if machine.CPU == nil {
			machine.CPU = &models.CPU{}
		}
		machine.CPU.Cores = int(config.Hardware.NumCPU)
		// disk and GPU
		updateDevice(machine, config.Hardware.Device)
	}

	if len(guest.IpStack) == 0 {
		logger.Warn("VM guest IP stack is empty")
		return
	}

	// TODO: update network
	updateNetwork(machine, networks, guest.IpStack)

}

func updateDevice(machine *models.Machine, devices []types.BaseVirtualDevice) {
	controllerKeys := make(map[int32]string)
	for _, dev := range devices {
		switch d := dev.(type) {
		case *types.ParaVirtualSCSIController:
		case *types.VirtualSCSIController:
			controllerKeys[d.Key] = "scsi"
		case *types.VirtualIDEController:
			controllerKeys[d.Key] = "ide"
		}
	}

	for _, dev := range devices {
		switch d := dev.(type) {
		case *types.VirtualDisk:
			description := d.DeviceInfo.GetDescription()
			controller := ""
			if value, ok := controllerKeys[d.ControllerKey]; ok {
				controller = value
			}
			size := uint64(0)
			if d.CapacityInBytes > 0 {
				size = uint64(d.CapacityInBytes)
			}
			disk := models.Disk{
				Name:       description.Label,
				Size:       size,
				Type:       getTypeFromBacking(d.Backing),
				Controller: controller,
			}

			for _, d := range machine.Disks {
				if d.Name == disk.Name {

					continue
				}
			}
			// here we suppoe that we do not have the disk yet
			machine.Disks = append(machine.Disks, &disk)
		case *types.VirtualMachineVideoCard:
			gpu := models.GPU{
				Index: int(*d.UnitNumber),
			}
			// here we suppoe that we do not have the gpu yet
			machine.GPUS = append(machine.GPUS, &gpu)
		default:
			// pass
		}
	}
}

func getTypeFromBacking(b types.BaseVirtualDeviceBackingInfo) string {
	switch b.(type) {
	case *types.VirtualDiskFlatVer2BackingInfo:
		return "vmdk"
	case *types.VirtualDiskRawDiskMappingVer1BackingInfo:
		return "raw"
	case *types.VirtualDiskRawDiskVer2BackingInfo:
		return "raw"
	case *types.VirtualDiskSparseVer2BackingInfo:
		return "sparse"
	case *types.VirtualDiskFlatVer1BackingInfo:
		return "flat"
	case *types.VirtualDiskSeSparseBackingInfo:
		return "se_sparse"
	case *types.VirtualDiskLocalPMemBackingInfo:
		return "pmem"
	case *types.VirtualDiskPartitionedRawDiskVer2BackingInfo:
		return "partitioned_raw"
	case *types.VirtualDiskSparseVer1BackingInfo:
		return "sparse_v1"
	default:
		return ""
	}
}

func updateNetwork(machine *models.Machine, networks []types.GuestNicInfo, ipstacks []types.GuestStackInfo) {
	var err error
	for _, network := range networks {
		var mac net.HardwareAddr
		var ip net.IP
		var nic *models.NetworkInterface

		if network.MacAddress != "" {
			mac, err = net.ParseMAC(network.MacAddress)
			if err == nil {
				nic = machine.GetNetworkInterfaceByMAC(mac)
			}
		}

		if nic == nil && len(network.IpConfig.IpAddress) > 0 {
			ip = net.ParseIP(network.IpAddress[0])
			if ip != nil {
				nic = machine.GetNetworkInterfaceByIP(ip)
			}
		}

		if nic == nil {
			nic = &models.NetworkInterface{MAC: mac}
			machine.NICS = append(machine.NICS, nic)
		}

		// update NIC
		for _, ipaddr := range network.IpConfig.IpAddress {
			parsedIP := net.ParseIP(ipaddr.IpAddress)
			if parsedIP != nil {
				if parsedIP.To4() != nil {
					nic.IP = parsedIP
					nic.MaskSize = int(ipaddr.PrefixLength)
				} else {
					nic.IP6 = parsedIP
					nic.Mask6Size = int(ipaddr.PrefixLength)
				}
			}

		}

		// update gateway
	stacksloop:
		for _, stack := range ipstacks {
			if stack.IpRouteConfig == nil {
				continue
			}
			for _, route := range stack.IpRouteConfig.IpRoute {
				gw := net.ParseIP(route.Gateway.IpAddress)
				if gw == nil {
					continue
				}
				_, nw, err := net.ParseCIDR(fmt.Sprintf("%s/%d", route.Network, route.PrefixLength))
				if err == nil {
					if nw.Contains(nic.IP) || nw.Contains(nic.IP6) {
						nic.Gateway = gw
						break stacksloop
					}
				}
			}

		}
	}
}
