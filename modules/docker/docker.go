package docker

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/store"
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
	return &Platform{machine: m, client: c}
}

func (p *Platform) Ping(ctx context.Context) error {
	_, err := p.client.Ping(ctx)
	return err
}

func createNetworkInterface(ipam network.IPAM, endpoint types.EndpointResource) *models.NetworkInterface {
	hw, err := net.ParseMAC(endpoint.MacAddress)
	if err != nil {
		fmt.Println(err)
		hw = net.HardwareAddr([]byte{0, 0, 0, 0, 0, 0})
	}
	nic := models.NetworkInterface{
		MAC:  hw,
		Name: endpoint.Name, // :warning: this is not the real network interface name in the container
	}

	ip4, subnet4, err := net.ParseCIDR(endpoint.IPv4Address)
	if err == nil {
		nic.IP = ip4
		nic.MaskSize, _ = subnet4.Mask.Size()
	}

	ip6, subnet6, err := net.ParseCIDR(endpoint.IPv6Address)
	if err == nil {
		nic.IP6 = ip6
		nic.Mask6Size, _ = subnet6.Mask.Size()
	}
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

func getOrCreateMachineFromEndpoint(
	endpoint types.EndpointResource,
	container types.Container,
	ipam network.IPAM,
	parent *models.Machine,
	logger *logrus.Entry) *models.Machine {
	machine := store.GetMachineByHostID(container.ID)
	// Otherwise, create it
	if machine == nil {
		image, version := splitImageName(container.Image)

		// machine
		machine = models.NewMachine()
		if len(container.Names) > 0 {
			machine.Hostname = strings.TrimPrefix(container.Names[0], "/") // use container name
		}
		machine.Platform = "docker"                                  // set platform to docker
		machine.Distribution = image                                 // container image
		machine.DistributionVersion = version                        // container image version
		machine.HostID = container.ID                                // container ID
		machine.Uptime = time.Since(time.Unix(container.Created, 0)) //
		machine.ParentMachine = parent.InternalID                    // underlying machine
		// fmt.Printf("%+v\n", machine)

		// logging
		logger.WithField("host_id", machine.HostID).
			WithField("hostname", machine.Hostname).
			WithField("uptime", machine.Uptime).
			WithField("platform", machine.Platform).
			WithField("distribution", machine.Distribution).
			WithField("distribution_version", machine.DistributionVersion).
			WithField("parent", machine.ParentMachine).
			Info("Create new container")

		store.InsertMachine(machine)
	}

	// check network information
	ip4 := net.ParseIP(endpoint.IPv4Address)
	ip6 := net.ParseIP(endpoint.IPv6Address)

	if machine.HasIP(ip4) && ip6 != nil {
		nic := machine.GetNetworkInterfaceByIP(ip4)
		nic.IP6 = ip6
		return machine
	}

	if machine.HasIP(ip6) && ip4 != nil {
		nic := machine.GetNetworkInterfaceByIP(ip4)
		nic.IP = ip4
		return machine
	}

	// otherwise create a network interface
	nic := createNetworkInterface(ipam, endpoint)
	machine.NICS = append(machine.NICS, nic)

	return machine
}

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

func getContainerByID(ctx context.Context, cli *client.Client, id string) (types.Container, error) {
	filters := filters.NewArgs()
	filters.Add("id", id)
	options := types.ContainerListOptions{
		Filters: filters,
	}
	containers, err := cli.ContainerList(ctx, options)
	if len(containers) == 1 {
		return containers[0], nil
	}
	return types.Container{}, err
}

func RunBasic(ctx context.Context, p *Platform, logger *logrus.Entry) error {
	// find all the networks
	networks, err := p.client.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return err
	}

	// loop over the networks
	for _, n := range networks {

		network, err := p.client.NetworkInspect(ctx, n.ID, types.NetworkInspectOptions{})
		// fmt.Println(network.Name, network.Containers)
		// keep only the used networks
		if err != nil || len(network.Containers) == 0 || len(network.IPAM.Config) == 0 {
			// go next
			continue
		}

		// loop over the containers
		for containerID, endpoint := range network.Containers {
			// fmt.Printf("\t - %v [%v] %s\n",
			// endpoint.IPv4Address,
			// endpoint.IPv6Address,
			// endpoint.Name)
			// container, err := m.getContainerByName(endpoint.Name)
			container, err := getContainerByID(ctx, p.client, containerID)
			// fmt.Printf(" ENDPOINT: %+v\n", endpoint)
			// fmt.Printf("CONTAINER: %+v\n\n", container)
			if err != nil || container.ID == "" {
				// in swarm mode (at least) there are some special containers attached
				// to the docker_gwbridge interface.
				continue
			}

			// ignore containers managed by docker swarm
			if id, exists := container.Labels["com.docker.swarm.service.id"]; exists && id != "" {
				logger.WithField("name", endpoint.Name).
					WithField("id", container.ID).
					Debugf("Find container managed by swarm (ignoring)")
				continue
			}

			machine := getOrCreateMachineFromEndpoint(endpoint, container, network.IPAM, p.machine, logger)

			apps := machine.Applications() // bypass the packages
			// here we have a machine
			// we must update its app
			var app *models.Application
			// get the app.
			// :warning: we consider a single app by container
			if len(apps) > 0 {
				app = apps[0]
			} else {
				app, _ = machine.GetOrCreateApplicationByName(machine.Distribution)
				// or create it
				// app = &models.Application{
				// 	Name: machine.Distribution,
				// 	// Version: machine.DistributionVersion,
				// }
				// machine.Applications = append(machine.Applications, app)
			}

			// add an endpoint for every exposed ports
			for _, port := range container.Ports {
				var ip net.IP
				if port.IP == "" {
					ip = net.IPv4zero
				} else if port.IP == "::" {
					ip = net.IPv6zero
				} else {
					ip = net.ParseIP(port.IP)
				}
				endpoint, created := app.AddEndpoint(ip, port.PrivatePort, port.Type)
				if created {
					logger.
						WithField("container", machine.Hostname).
						WithField("ip", endpoint.Addr).
						WithField("port", endpoint.Port).
						WithField("proto", endpoint.Protocol).
						Info("Application endpoint found")
				}
			}

		}
	}
	return nil
}
