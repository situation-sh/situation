package docker

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"

	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/store"
	"github.com/situation-sh/situation/pkg/utils"
)

var Zero4 = netip.AddrFrom4([4]byte{0, 0, 0, 0})
var Zero6 = netip.AddrFrom16([16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})

// Platform is a struct that aims to
// communicate with a docker instance. The
// machine attribute helps to associate the
// docker instance with the underlying machine
type Platform struct {
	machine *models.Machine
	client  *client.Client
}

func NewPlatform(m *models.Machine, c *client.Client) *Platform {
	// c.NegotiateAPIVersion(context.Background())
	return &Platform{machine: m, client: c}
}

func (p *Platform) Ping(ctx context.Context) error {
	_, err := p.client.Ping(ctx, client.PingOptions{NegotiateAPIVersion: true})
	return err
}

// func createNetworkInterface(ipam network.IPAM, endpoint network.EndpointResource) *models.NetworkInterface {
// 	hw, err := net.ParseMAC(endpoint.MacAddress)
// 	if err != nil {
// 		// fmt.Println(err)
// 		hw = net.HardwareAddr([]byte{0, 0, 0, 0, 0, 0})
// 	}
// 	nic := models.NetworkInterface{
// 		MAC:  hw.String(),
// 		Name: endpoint.Name, // :warning: this is not the real network interface name in the container
// 		IP:   make([]string, 0),
// 	}

// 	// ip4, subnet4, err := net.ParseCIDR(endpoint.IPv4Address)
// 	// if err == nil {
// 	// 	nic.IP = ip4
// 	// 	nic.MaskSize, _ = subnet4.Mask.Size()
// 	// }
// 	if endpoint.IPv4Address != "" {
// 		nic.IP = append(nic.IP, endpoint.IPv4Address)
// 	}
// 	if endpoint.IPv6Address != "" {
// 		nic.IP = append(nic.IP, endpoint.IPv6Address)
// 	}

// 	// ip6, subnet6, err := net.ParseCIDR(endpoint.IPv6Address)
// 	// if err == nil {
// 	// 	nic.IP6 = ip6
// 	// 	nic.Mask6Size, _ = subnet6.Mask.Size()
// 	// }
// 	return &nic
// }

func splitImageName(image string) (string, string) {
	// parse image and version
	s := strings.Split(image, ":")
	if len(s) <= 1 {
		return s[0], "latest"
	}
	// sometimes there is a hash in the version: version@sha512:...
	version := strings.Split(s[1], "@")[0]
	return s[0], version

}

// func getOrCreateMachineFromEndpoint(
// 	endpoint network.EndpointResource,
// 	container container.InspectResponse,
// 	ipam network.IPAM,
// 	parent *models.Machine,
// 	logger logrus.FieldLogger,
// 	s *store.BunStorage) (*models.Machine, *models.NetworkInterface) {
// 	machine := s.GetMachineByHostID(container.ID)
// 	// Otherwise, create it
// 	if machine == nil {
// 		image, version := splitImageName(container.Config.Image)

// 		uptime := time.Duration(-1)
// 		createdAt, err := time.Parse(time.RFC3339, container.Created)
// 		if err == nil {
// 			// here we have the right uptime (int64 -> ns)
// 			uptime = time.Since(createdAt)
// 		}
// 		// machine
// 		// machine = models.NewMachine()
// 		// if len(container.Names) > 0 {
// 		// 	machine.Hostname = strings.TrimPrefix(container.Names[0], "/") // use container name
// 		// }
// 		machine = &models.Machine{
// 			Hostname:            strings.TrimPrefix(container.Name, "/"),
// 			Platform:            "docker",
// 			Distribution:        image,
// 			DistributionVersion: version,
// 			HostID:              container.ID,
// 			Uptime:              uptime,
// 			ParentMachine:       parent,
// 			ParentMachineID:     parent.ID,
// 		}
// 		// machine.Hostname = strings.TrimPrefix(container.Name, "/")
// 		// machine.Platform = "docker"           // set platform to docker
// 		// machine.Distribution = image          // container image
// 		// machine.DistributionVersion = version // container image version
// 		// machine.HostID = container.ID         // container ID
// 		// machine.Uptime = uptime               //
// 		// machine.ParentMachine = parent        // underlying machine
// 		// machine.ParentMachineID = parent.ID
// 		// fmt.Printf("%+v\n", machine)

// 		// logging
// 		logger.WithField("host_id", machine.HostID).
// 			WithField("hostname", machine.Hostname).
// 			WithField("uptime", machine.Uptime).
// 			WithField("platform", machine.Platform).
// 			WithField("distribution", machine.Distribution).
// 			WithField("distribution_version", machine.DistributionVersion).
// 			WithField("parent", machine.ParentMachineID).
// 			Info("Create new container")

// 		// s.InsertMachine(machine)
// 	}

// 	var nic *models.NetworkInterface
// 	if len(machine.NICS) == 0 {
// 		nic = &models.NetworkInterface{}
// 	} else {
// 		nic = machine.NICS[0]
// 	}

// 	// check network information
// 	ip4 := net.ParseIP(endpoint.IPv4Address)
// 	ip6 := net.ParseIP(endpoint.IPv6Address)

// 	if machine.HasIP(ip4) && ip6 != nil {
// 		nic := machine.GetNetworkInterfaceByIP(ip4)
// 		nic.IP6 = ip6
// 		return machine
// 	}

// 	if machine.HasIP(ip6) && ip4 != nil {
// 		nic := machine.GetNetworkInterfaceByIP(ip4)
// 		nic.IP = ip4
// 		return machine
// 	}

// 	// otherwise create a network interface
// 	// nic := createNetworkInterface(ipam, endpoint)
// 	machine.NICS = append(machine.NICS, nic)

// 	return machine
// }

// func getContainerByName(cli *client.Client, ctx context.Context, name string) (types.Container, error) {
// 	filters := filters.NewArgs()
// 	filters.Add("name", name)
// 	options := types.ContainerListOptions{
// 		Filters: filters,
// 	}
// 	containers, err := cli.ContainerList(ctx, options)
// 	if len(containers) == 1 {
// 		return containers[0], nil
// 	}
// 	return types.Container{}, err
// }

func getContainerByID(ctx context.Context, cli *client.Client, id string) (container.Summary, error) {
	filters := client.Filters{}
	filters.Add("id", id)
	// options := container.ListOptions{
	// 	Filters: filters,
	// }
	options := client.ContainerListOptions{Filters: filters}
	containers, err := cli.ContainerList(ctx, options)
	if len(containers.Items) == 1 {
		return containers.Items[0], nil
	}
	return container.Summary{}, err
}

func tagSubnets(ctx context.Context, p *Platform, logger logrus.FieldLogger, s *store.BunStorage) error {
	hostID := s.GetHostID(ctx)
	if hostID <= 0 {
		return fmt.Errorf("host not found in storage")
	}
	filters := client.Filters{}
	filters.Add("driver", "bridge")
	options := client.NetworkListOptions{Filters: filters}
	// options := network.ListOptions{Filters: filters.NewArgs(filters.Arg("driver", "bridge"))}
	networks, err := p.client.NetworkList(ctx, options)
	if err != nil {
		return err
	}

	subnets := make([]*models.Subnetwork, 0)

	// loop over the networks
	for _, n := range networks.Items {
		network, err := p.client.NetworkInspect(ctx, n.ID, client.NetworkInspectOptions{})
		if err != nil {
			logger.WithError(err).WithField("network_id", n.ID).Warn("fail to inspect network")
			continue
		}
		iface, ok := network.Network.Options["com.docker.network.bridge.name"]
		if ok {
			var link models.NetworkInterfaceSubnet
			err = s.DB().
				NewSelect().
				Model(&link).
				Relation("Subnetwork").
				Relation("NetworkInterface").
				Where("network_interface_id IN (?)",
					s.DB().NewSelect().
						Model((*models.NetworkInterface)(nil)).
						Column("id").
						Where("name = ?", iface).
						Where("machine_id = ?", hostID),
				).
				Limit(1).
				Scan(ctx)
			if err != nil {
				logger.
					WithError(err).
					WithField("iface", iface).
					WithField("host_id", hostID).
					Warn("no host iface found")
				continue
			}
			if link.Subnetwork != nil {
				link.Subnetwork.Tag = network.Network.ID
				// set the gateway if there is only one config
				// (most of the time it is the case for bridge networks)
				if len(network.Network.IPAM.Config) == 1 {
					link.Subnetwork.Gateway = network.Network.IPAM.Config[0].Gateway.String()
				}
				subnets = append(subnets, link.Subnetwork)
			}
		}
	}

	if len(subnets) > 0 {
		_, err = s.DB().
			NewUpdate().
			Model(&subnets).
			Bulk().
			Exec(ctx)
		if err != nil {
			logger.
				WithError(err).
				WithField("subnets", len(subnets)).
				Warn("failed to update subnet tags")
		}
		logger.WithField("subnets", len(subnets)).Info("Subnet tags updated")
	}

	return nil
}

func RunBasic(ctx context.Context, p *Platform, logger logrus.FieldLogger, s *store.BunStorage) error {
	// tag subnets with docker network id
	if err := tagSubnets(ctx, p, logger, s); err != nil {
		return fmt.Errorf("failed to tag subnets: %v", err)
	}
	// find all the networks
	networks, err := p.client.NetworkList(ctx, client.NetworkListOptions{})
	if err != nil {
		return err
	}

	// loop over the networks
	for _, n := range networks.Items {

		network, err := p.client.NetworkInspect(ctx, n.ID, client.NetworkInspectOptions{})
		// fmt.Println(network.Name, network.Containers)
		// keep only the used networks
		if err != nil || len(network.Network.Containers) == 0 || len(network.Network.IPAM.Config) == 0 {
			// go next
			continue
		}

		// create subnets
		subnetMap := make(map[string]*models.Subnetwork)
		subnets := make([]*models.Subnetwork, 0)
		for _, cfg := range network.Network.IPAM.Config {

			addr, cidr, err := net.ParseCIDR(cfg.Subnet.String())
			if err != nil {
				logger.
					WithField("network_id", network.Network.ID).
					WithField("subnet", cfg.Subnet).
					WithError(err).
					Warn("fail to parse network")
				continue
			}

			// if network.Driver == "bridge" {
			// 	iface, ok := network.Options["com.docker.network.bridge.name"]
			// 	if ok {
			// 		storage.D
			// 	}

			// }
			// TODO: here we may have a problem since the docker network can
			// already be discovered by the host-network module.
			subnet := models.Subnetwork{
				NetworkAddr: addr.String(),
				NetworkCIDR: cidr.String(),
				Gateway:     cfg.Gateway.String(),
				MaskSize:    utils.MaskSize(cidr),
				IPVersion:   utils.IPVersion(addr),
				Tag:         network.Network.ID, // we use the docker network id as extra tag
			}

			subnets = append(subnets, &subnet)
			subnetMap[network.Network.ID] = &subnet
		}

		// upserts subnetworks
		err = s.DB().NewInsert().
			Model(&subnets).
			On("CONFLICT (network_cidr, tag) DO UPDATE").
			Set("updated_at = CURRENT_TIMESTAMP").
			Scan(ctx)
		if err != nil {
			logger.
				WithError(err).
				WithField("subnetworks", len(subnets)).
				WithField("network_id", network.Network.ID).
				Warn("failed to insert subnetworks")
			continue
		}

		// loop over the containers
		for containerID, endpoint := range network.Network.Containers {
			container, err := getContainerByID(ctx, p.client, containerID)

			if err != nil || container.ID == "" {
				// in swarm mode (at least) there are some special containers attached
				// to the docker_gwbridge interface.
				continue
			}

			// ignore containers managed by docker swarm
			if id, exists := container.Labels["com.docker.swarm.service.id"]; exists && id != "" {
				logger.
					WithField("name", endpoint.Name).
					WithField("id", container.ID).
					Debug("Find container managed by swarm (ignoring)")
				continue
			}

			// we prefer inspect because of the startedAt property (=uptime)
			// that is easier to get
			containerJSON, err := p.client.ContainerInspect(ctx, container.ID, client.ContainerInspectOptions{})
			if err != nil {
				// TODO -> LOG

				// normally we can't be here because the container exist.
				// If something goes wrong just continue
				continue
			}

			// get or create the corresponding machine
			image, version := splitImageName(containerJSON.Container.Config.Image)
			uptime := time.Duration(0)
			createdAt, err := time.Parse(time.RFC3339, containerJSON.Container.Created)
			if err == nil {
				// here we have the right uptime (int64 -> ns)
				uptime = time.Since(createdAt)
			}
			machine := models.Machine{
				Hostname:            strings.TrimPrefix(containerJSON.Container.Name, "/"),
				Platform:            "docker",
				Distribution:        image,
				DistributionVersion: version,
				HostID:              container.ID,
				Uptime:              uptime,
				ParentMachine:       p.machine,
				ParentMachineID:     p.machine.ID,
				Chassis:             "container",
			}
			err = s.DB().
				NewInsert().
				Model(&machine).
				On("CONFLICT (host_id) DO UPDATE").
				Set("hostname = EXCLUDED.hostname").
				Set("platform = EXCLUDED.platform").
				Set("distribution = EXCLUDED.distribution").
				Set("distribution_version = EXCLUDED.distribution_version").
				Set("uptime = EXCLUDED.uptime").
				Set("parent_machine_id = EXCLUDED.parent_machine_id").
				Set("updated_at = CURRENT_TIMESTAMP").
				Scan(ctx)
			if err != nil {
				logger.WithError(err).
					WithField("container_name", containerJSON.Container.Name).
					WithField("container_id", containerJSON.Container.ID).
					WithField("image", machine.Hostname).
					WithField("distribution", machine.Distribution).
					WithField("distribution_version", machine.DistributionVersion).
					Warn("fail to create machine")
			}
			logger.
				WithField("container_name", containerJSON.Container.Name).
				WithField("container_id", containerJSON.Container.ID).
				WithField("image", machine.Hostname).
				WithField("distribution", machine.Distribution).
				WithField("distribution_version", machine.DistributionVersion).
				Info("created or updated machine")

			// machine nic
			settings, exists := containerJSON.Container.NetworkSettings.Networks[network.Network.Name]
			if !exists {
				logger.
					WithField("network", network.Network.Name).
					WithField("container_name", containerJSON.Container.Name).
					WithField("container_id", containerJSON.Container.ID).
					WithField("endpoint_id", endpoint.EndpointID).
					WithField("mac", endpoint.MacAddress).
					WithField("ipv4", endpoint.IPv4Address).
					Warn("cannot access container network settings")
				continue
			}

			nic := models.NetworkInterface{
				MAC:       settings.MacAddress.String(),
				IP:        make([]string, 0),
				Gateway:   settings.Gateway.String(),
				Machine:   &machine,
				MachineID: machine.ID,
				Flags:     models.NetworkInterfaceFlags{Up: true},
				Tag:       endpoint.EndpointID,
			}

			empty := netip.Addr{}
			if settings.IPAddress != empty {
				nic.IP = append(nic.IP, settings.IPAddress.String())
			}
			if settings.GlobalIPv6Address != empty {
				nic.IP = append(nic.IP, settings.GlobalIPv6Address.String())
			}

			err = s.DB().
				NewInsert().
				Model(&nic).
				On("CONFLICT (machine_id, mac, tag) DO UPDATE").
				Set("ip = EXCLUDED.ip").
				Set("gateway = EXCLUDED.gateway").
				Set("updated_at = CURRENT_TIMESTAMP").
				Scan(ctx)
			if err != nil {
				logger.
					WithError(err).
					WithField("container_name", containerJSON.Container.Name).
					WithField("mac", nic.MAC).
					WithField("ip", nic.IP).
					WithField("gateway", nic.Gateway).
					Warn("fail to create network interface")
				continue
			}
			logger.
				WithField("container_name", containerJSON.Container.Name).
				WithField("mac", nic.MAC).
				WithField("ip", nic.IP).
				WithField("gateway", nic.Gateway).
				Info("created or updated network interface")

			// link nic to subnetwork
			subnet, exists := subnetMap[network.Network.ID]
			if exists {
				link := models.NetworkInterfaceSubnet{
					NetworkInterface:   &nic,
					NetworkInterfaceID: nic.ID,
					Subnetwork:         subnet,
					SubnetworkID:       subnet.ID,
				}
				if err := link.SetMACSubnet(); err != nil {
					logger.
						WithError(err).
						WithField("container_name", containerJSON.Container.Name).
						WithField("mac", nic.MAC).
						WithField("subnetwork_id", subnet.ID).
						Warn("fail to set MACSubnet for network interface subnet link")
					continue
				}
				_, err = s.DB().
					NewInsert().
					Model(&link).
					On("CONFLICT (mac_subnet) DO NOTHING").
					Exec(ctx)
				if err != nil {
					logger.
						WithError(err).
						WithField("container_name", containerJSON.Container.Name).
						WithField("mac", nic.MAC).
						WithField("subnetwork_id", subnet.ID).
						Warn("fail to link network interface to subnetwork")
				} else {
					logger.
						WithField("container_name", containerJSON.Container.Name).
						WithField("mac", nic.MAC).
						WithField("subnetwork_id", subnet.ID).
						Info("linked network interface to subnetwork")
				}
			}

			// now create ports
			endpoints := make([]*models.ApplicationEndpoint, 0)
			policies := make([]*models.EndpointPolicy, 0)
			for _, port := range container.Ports {
				// container endpoint
				ctrEndpoint := models.ApplicationEndpoint{
					Port:               uint16(port.PrivatePort),
					Protocol:           port.Type,
					NetworkInterfaceID: nic.ID,
					NetworkInterface:   &nic,
				}
				endpoints = append(endpoints, &ctrEndpoint)

				// TODO fix error: (we need a kind of hash i think)
				// WRN docker fail to create application endpoints container_name="/nostalgic_galois" endpoints="9" error="ERROR: ON CONFLICT DO UPDATE command cannot affect row a second time (SQLSTATE=21000)" ip="[172.17.0.2]"
				for _, hostNIC := range p.machine.NICS {
					for _, addr := range hostNIC.IP {
						if (port.IP == Zero4 && utils.IPVersionString(addr) == 4) ||
							(port.IP == Zero6 && utils.IPVersionString(addr) == 6) ||
							(port.IP.String() == addr) {
							// host endpoint
							hostEndpoint := models.ApplicationEndpoint{
								Addr:               addr,
								Port:               uint16(port.PublicPort),
								Protocol:           port.Type,
								NetworkInterfaceID: hostNIC.ID,
								NetworkInterface:   hostNIC,
							}
							endpoints = append(endpoints, &hostEndpoint)

							// forward rule from host endpoint to container endpoint
							policy := models.EndpointPolicy{
								Endpoint:    &ctrEndpoint,
								Action:      "forward",
								SrcEndpoint: &hostEndpoint,
								Source:      "docker",
							}
							policies = append(policies, &policy)
							break
						}
					}
				}
				// }
				// fmt.Printf("Port: %+v\n", port)
			}
			// 	for _, hostNIC := range p.machine.NICS {
			// 		for _, addr := range hostNIC.IP {
			// 			if addr == port.IP {
			// 				// need to get the local endpoint
			// 				src := models.ApplicationEndpoint{
			// 					Addr:               addr,
			// 					Port:               uint16(port.PublicPort),
			// 					Protocol:           port.Type,
			// 					NetworkInterfaceID: hostNIC.ID,
			// 					NetworkInterface:   hostNIC,
			// 				}
			// 				endpoints = append(endpoints, &src)

			// 				policy := models.EndpointPolicy{
			// 					Endpoint:    &endpoint,
			// 					Action:      "forward",
			// 					SrcEndpoint: &src,
			// 					Source:      "docker",
			// 				}
			// 				policies = append(policies, &policy)
			// 				break
			// 			}
			// 		}
			// 	}
			// }

			if len(endpoints) > 0 {
				err = s.DB().
					NewInsert().
					Model(&endpoints).
					On("CONFLICT (network_interface_id, addr, port, protocol) DO UPDATE").
					Set("updated_at = CURRENT_TIMESTAMP").
					Scan(ctx)
				if err != nil {
					logger.
						WithError(err).
						WithField("container_name", containerJSON.Container.Name).
						WithField("ip", nic.IP).
						WithField("endpoints", len(endpoints)).
						Warn("fail to create application endpoints")
					continue
				}
				logger.
					WithField("container_name", containerJSON.Container.Name).
					WithField("ip", nic.IP).
					WithField("endpoints", len(endpoints)).
					Info("created or updated application endpoints")

				if len(policies) > 0 {
					for _, p := range policies {
						p.EndpointID = p.Endpoint.ID
						p.SrcEndpointID = p.SrcEndpoint.ID
					}
					_, err = s.DB().
						NewInsert().
						Model(&policies).
						On("CONFLICT (endpoint_id, action, src_endpoint_id, src_addr) DO UPDATE").
						Set("updated_at = CURRENT_TIMESTAMP").
						Exec(ctx)
					if err != nil {
						logger.
							WithError(err).
							WithField("container_name", containerJSON.Container.Name).
							WithField("ip", nic.IP).
							WithField("policies", len(policies)).
							Warn("fail to create endpoint policies")
						continue
					}
					logger.
						WithField("container_name", containerJSON.Container.Name).
						WithField("ip", nic.IP).
						WithField("policies", len(policies)).
						Info("created or updated endpoint policies")
				} else {
					logger.
						WithField("container_name", containerJSON.Container.Name).
						Warn("no endpoint policies to insert")
				}

			} else {
				logger.
					WithField("container_name", containerJSON.Container.Name).
					WithField("ip", nic.IP).
					Warn("no application endpoint to insert")
			}

		}
	}
	return nil
}
