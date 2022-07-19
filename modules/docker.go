package modules

import (
	"context"
	"fmt"
	"runtime"

	"github.com/docker/docker/client"
	docker "github.com/situation-sh/situation/modules/docker"
	"github.com/situation-sh/situation/store"
)

const defaultDockerPort = 2376

func init() {
	RegisterModule(&DockerModule{})
}

// Module definition ---------------------------------------------------------

type DockerModule struct{}

func (m *DockerModule) Name() string {
	return "docker"
}

func (m *DockerModule) Dependencies() []string {
	return []string{"host-network", "tcp-scan"}
}

func (m *DockerModule) Run() error {
	logger := GetLogger(m)
	ctx := context.Background()

	for _, platform := range m.findDockerInstances() {
		if err := platform.Ping(ctx); err != nil {
			logger.Warn(err)
			continue
		}
		if err := docker.RunBasic(ctx, platform, logger); err != nil {
			logger.Warn(err)
		}
		if err := docker.RunSwarm(ctx, platform, logger); err != nil {
			logger.Warn(err)
		}
	}

	return nil
}

// findDockerInstances return all the working docker clients. It means
// that it returns a way to communicate to all these docker instances.
func (m *DockerModule) findDockerInstances() []*docker.Platform {
	logger := GetLogger(m)
	hostMachine := store.GetHost()
	out := make([]*docker.Platform, 0)

	// try locally
	// -> first from env
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		logger.Debugf("docker instance not found from environment: %v. Trying manually...", err)
		// now try with default values
		host := "unix:///var/run/docker.sock"
		if runtime.GOARCH == "windows" {
			host = "tcp://127.0.0.1:2376"
		}
		cli, err = client.NewClientWithOpts(client.WithHost(host))
		if err != nil {
			logger.Debugf("docker instance not found with default values: %v", err)
		} else {
			out = append(out, docker.NewPlatform(hostMachine, cli))
		}
	} else {
		out = append(out, docker.NewPlatform(hostMachine, cli))
	}

	// try from network
	machines, endpoints := store.GetMachinesByOpenTCPPort(defaultDockerPort)
	for k, machine := range machines {
		if machine == hostMachine {
			// ignore the candidate that could be the current
			// host
			continue
		}
		host := fmt.Sprintf("tcp://%v:%d", endpoints[k].Addr, defaultDockerPort)
		cli, err = client.NewClientWithOpts(client.WithHost(host))
		if err != nil {
			logger.Debugf("trying to connect to %s fails: %v", host, err)
			continue
		}
		logger.Debugf("docker instance found at %s", host)
		out = append(out, docker.NewPlatform(machine, cli))
	}
	return out
}
