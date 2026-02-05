package docker

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/store"
	"github.com/situation-sh/situation/pkg/utils"
)

// Platform is a struct that aims to
// communicate with a docker instance. The
// machine attribute helps to associate the
// docker instance with the underlying machine
type Platform struct {
	machine *models.Machine
	client  *client.Client
}

func NewPlatform(m *models.Machine, c *client.Client) *Platform {
	c.NegotiateAPIVersion(context.Background())
	return &Platform{machine: m, client: c}
}

func (p *Platform) Ping(ctx context.Context) error {
	_, err := p.client.Ping(ctx)
	return err
}

func createNetworkInterface(ipam network.IPAM, endpoint network.EndpointResource) *models.NetworkInterface {
	hw, err := net.ParseMAC(endpoint.MacAddress)
	if err != nil {
		fmt.Println(err)
		hw = net.HardwareAddr([]byte{0, 0, 0, 0, 0, 0})
	}
	nic := models.NetworkInterface{
		MAC:  hw.String(),
		Name: endpoint.Name, // :warning: this is not the real network interface name in the container
		IP:   make([]string, 0),
	}

	// ip4, subnet4, err := net.ParseCIDR(endpoint.IPv4Address)
	// if err == nil {
	// 	nic.IP = ip4
	// 	nic.MaskSize, _ = subnet4.Mask.Size()
	// }
	if endpoint.IPv4Address != "" {
		nic.IP = append(nic.IP, endpoint.IPv4Address)
	}
	if endpoint.IPv6Address != "" {
		nic.IP = append(nic.IP, endpoint.IPv6Address)
	}

	// ip6, subnet6, err := net.ParseCIDR(endpoint.IPv6Address)
	// if err == nil {
	// 	nic.IP6 = ip6
	// 	nic.Mask6Size, _ = subnet6.Mask.Size()
	// }
	return &nic
}

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
	filters := filters.NewArgs()
	filters.Add("id", id)
	options := container.ListOptions{
		Filters: filters,
	}
	containers, err := cli.ContainerList(ctx, options)
	if len(containers) == 1 {
		return containers[0], nil
	}
	return container.Summary{}, err
}

func RunBasic(ctx context.Context, p *Platform, logger logrus.FieldLogger, s *store.BunStorage) error {
	// find all the networks
	networks, err := p.client.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return err
	}

	// loop over the networks
	for _, n := range networks {

		network, err := p.client.NetworkInspect(ctx, n.ID, network.InspectOptions{})
		// fmt.Println(network.Name, network.Containers)
		// keep only the used networks
		if err != nil || len(network.Containers) == 0 || len(network.IPAM.Config) == 0 {
			// go next
			continue
		}

		// create subnets
		subnetMap := make(map[string]*models.Subnetwork)
		subnets := make([]*models.Subnetwork, 0)
		for _, cfg := range network.IPAM.Config {
			addr, cidr, err := net.ParseCIDR(cfg.Subnet)
			if err != nil {
				logger.
					WithField("network_id", network.ID).
					WithField("subnet", cfg.Subnet).
					WithError(err).
					Warn("fail to parse network")
				continue
			}
			subnet := models.Subnetwork{
				NetworkAddr: addr.String(),
				NetworkCIDR: cidr.String(),
				Gateway:     cfg.Gateway,
				MaskSize:    utils.MaskSize(cidr),
				IPVersion:   utils.IPVersion(addr),
				Tag:         network.ID, // we use the docker network id as extra tag
			}

			subnets = append(subnets, &subnet)
			subnetMap[network.ID] = &subnet
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
				WithField("network_id", network.ID).
				Warn("failed to insert subnetworks")
			continue
		}

		// loop over the containers
		for containerID, endpoint := range network.Containers {
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
			containerJSON, err := p.client.ContainerInspect(ctx, container.ID)
			if err != nil {
				// TODO -> LOG

				// normally we can't be here because the container exist.
				// If something goes wrong just continue
				continue
			}

			// get or create the corresponding machine
			image, version := splitImageName(containerJSON.Config.Image)
			uptime := time.Duration(0)
			createdAt, err := time.Parse(time.RFC3339, containerJSON.Created)
			if err == nil {
				// here we have the right uptime (int64 -> ns)
				uptime = time.Since(createdAt)
			}
			machine := models.Machine{
				Hostname:            strings.TrimPrefix(containerJSON.Name, "/"),
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
					WithField("container_name", containerJSON.Name).
					WithField("container_id", containerJSON.ID).
					WithField("image", machine.Hostname).
					WithField("distribution", machine.Distribution).
					WithField("distribution_version", machine.DistributionVersion).
					Warn("fail to create machine")
			}
			logger.
				WithField("container_name", containerJSON.Name).
				WithField("container_id", containerJSON.ID).
				WithField("image", machine.Hostname).
				WithField("distribution", machine.Distribution).
				WithField("distribution_version", machine.DistributionVersion).
				Info("created or updated machine")

			// machine nic
			settings, exists := containerJSON.NetworkSettings.Networks[network.Name]
			if !exists {
				logger.
					WithField("network", network.Name).
					WithField("container_name", containerJSON.Name).
					WithField("container_id", containerJSON.ID).
					WithField("endpoint_id", endpoint.EndpointID).
					WithField("mac", endpoint.MacAddress).
					WithField("ipv4", endpoint.IPv4Address).
					Warn("cannot access container network settings")
				continue
			}

			nic := models.NetworkInterface{
				MAC:       settings.MacAddress,
				IP:        make([]string, 0),
				Gateway:   settings.Gateway,
				Machine:   &machine,
				MachineID: machine.ID,
				Tag:       endpoint.EndpointID,
			}

			if settings.IPAddress != "" {
				nic.IP = append(nic.IP, settings.IPAddress)
			}
			if settings.GlobalIPv6Address != "" {
				nic.IP = append(nic.IP, settings.GlobalIPv6Address)
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
					WithField("container_name", containerJSON.Name).
					WithField("mac", nic.MAC).
					WithField("ip", nic.IP).
					WithField("gateway", nic.Gateway).
					Warn("fail to create network interface")
				continue
			}
			logger.
				WithField("container_name", containerJSON.Name).
				WithField("mac", nic.MAC).
				WithField("ip", nic.IP).
				WithField("gateway", nic.Gateway).
				Info("created or updated network interface")

			// link nic to subnetwork
			subnet, exists := subnetMap[network.ID]
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
						WithField("container_name", containerJSON.Name).
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
						WithField("container_name", containerJSON.Name).
						WithField("mac", nic.MAC).
						WithField("subnetwork_id", subnet.ID).
						Warn("fail to link network interface to subnetwork")
				} else {
					logger.
						WithField("container_name", containerJSON.Name).
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

				for _, hostNIC := range p.machine.NICS {
					for _, addr := range hostNIC.IP {
						if (port.IP == "0.0.0.0" && utils.IPVersionString(addr) == 4) ||
							(port.IP == "::" && utils.IPVersionString(addr) == 6) ||
							(port.IP == addr) {
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
				fmt.Printf("Port: %+v\n", port)
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
						WithField("container_name", containerJSON.Name).
						WithField("ip", nic.IP).
						WithField("endpoints", len(endpoints)).
						Warn("fail to create application endpoints")
					continue
				}
				logger.
					WithField("container_name", containerJSON.Name).
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
							WithField("container_name", containerJSON.Name).
							WithField("ip", nic.IP).
							WithField("policies", len(policies)).
							Warn("fail to create endpoint policies")
						continue
					}
					logger.
						WithField("container_name", containerJSON.Name).
						WithField("ip", nic.IP).
						WithField("policies", len(policies)).
						Info("created or updated endpoint policies")
				} else {
					logger.
						WithField("container_name", containerJSON.Name).
						Warn("no endpoint policies to insert")
				}

			} else {
				logger.
					WithField("container_name", containerJSON.Name).
					WithField("ip", nic.IP).
					Warn("no application endpoint to insert")
			}

		}
	}
	return nil
}
