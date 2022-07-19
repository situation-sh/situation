package docker

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/store"
)

func RunSwarm(ctx context.Context, p *Platform, logger *logrus.Entry) error {
	services, err := p.client.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return err
	}
	for _, service := range services {

		// fmt.Printf("Name: %+v\n", service.Spec.Name)
		// fmt.Printf("ID: %+v\n", service.ID)
		// fmt.Printf("Meta: %+v\n", service.Meta)
		// fmt.Printf("Networks: %+v\n", service.Spec.Networks)
		// fmt.Printf("Task: %+v\n", service.Spec.TaskTemplate)
		// fmt.Printf("Endpoint: %+v\n", service.Endpoint)
		// fmt.Printf("Network: %+v\n\n", service.Endpoint.VirtualIPs)

		machine := store.GetMachineByHostID(service.ID)
		if machine == nil {
			image, version := splitImageName(service.Spec.TaskTemplate.ContainerSpec.Image)
			machine = models.NewMachine()

			machine.Hostname = service.Spec.Name           // use container name
			machine.Platform = "swarm"                     // set platform to docker
			machine.Distribution = image                   // container image
			machine.DistributionVersion = version          // container image version
			machine.HostID = service.ID                    // container ID
			machine.Uptime = time.Since(service.CreatedAt) //
			machine.ParentMachine = p.machine              // underlying machine

			// logging
			logger.WithField("host_id", machine.HostID).
				WithField("hostname", machine.Hostname).
				WithField("uptime", machine.Uptime).
				WithField("platform", machine.Platform).
				WithField("distribution", machine.Distribution).
				WithField("distribution_version", machine.DistributionVersion).
				WithField("parent", machine.ParentMachine.HostID).
				Info("Create new swarm service")

			store.InsertMachine(machine)
		}
	}
	return nil
}
