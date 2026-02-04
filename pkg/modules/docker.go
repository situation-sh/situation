// LINUX(DockerModule) ok
// WINDOWS(DockerModule) ok
// MACOS(DockerModule) ?
// ROOT(DockerModule) yes
package modules

import (
	"context"
	"runtime"

	"github.com/asiffer/puzzle"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	docker "github.com/situation-sh/situation/pkg/modules/docker"
	"github.com/situation-sh/situation/pkg/store"
)

const defaultDockerPort = 2376

func init() {
	registerModule(&DockerModule{host: defaultDockerHost()})
}

func defaultDockerHost() string {
	if runtime.GOOS == "windows" {
		return "tcp://127.0.0.1:2376"
	}
	return "unix:///var/run/docker.sock"
}

// Module definition ---------------------------------------------------------

// DockerModule retrieves information about docker containers.
//
// It uses the official go client that performs HTTP queries
// either on port `:2375` (on windows generally) or on UNIX sockets.
//
// We generally need some privileges to reads UNIX sockets, so it may
// require root privileges (the alternative is to belong to the `docker` group)
type DockerModule struct {
	BaseModule

	host string
}

func (m *DockerModule) Bind(config *puzzle.Config) error {
	if err := setDefault(config, m, "host", &m.host, "Local docker host to scan after reading from env"); err != nil {
		return err
	}
	return nil
}

func (m *DockerModule) Name() string {
	return "docker"
}

func (m *DockerModule) Dependencies() []string {
	return []string{"host-network", "tcp-scan"}
}

func (m *DockerModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	for _, platform := range m.findDockerInstances(ctx, logger, storage) {
		if err := platform.Ping(ctx); err != nil {
			logger.Warn(err)
			continue
		}
		if err := docker.RunBasic(ctx, platform, logger, storage); err != nil {
			logger.Warn(err)
		}
		// if err := docker.RunSwarm(ctx, platform, logger, storage); err != nil {
		// 	logger.Warn(err)
		// }
	}

	return nil
}

// findDockerInstances return all the working docker clients. It means
// that it returns a way to communicate to all these docker instances.
func (m *DockerModule) findDockerInstances(ctx context.Context, logger logrus.FieldLogger, s *store.BunStorage) []*docker.Platform {

	hostMachine := s.GetHost(ctx)
	out := make([]*docker.Platform, 0)

	// try locally
	// -> first from env
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		logger.
			WithError(err).
			Debug("instance not found from environment")
		cli, err = client.NewClientWithOpts(client.WithHost(m.host))
		if err != nil {
			logger.
				WithError(err).
				Warn("instance not found")
		} else {
			out = append(out, docker.NewPlatform(hostMachine, cli))
		}
	} else {
		out = append(out, docker.NewPlatform(hostMachine, cli))
	}

	// try from network
	// machines, _, endpoints := s.GetMachinesByOpenTCPPort(defaultDockerPort)
	// for k, machine := range machines {
	// 	if machine == hostMachine {
	// 		// ignore the candidate that could be the current
	// 		// host
	// 		continue
	// 	}
	// 	host := fmt.Sprintf("tcp://%v:%d", endpoints[k].Addr, defaultDockerPort)
	// 	cli, err = client.NewClientWithOpts(client.WithHost(host))
	// 	if err != nil {
	// 		logger.Debugf("trying to connect to %s fails: %v", host, err)
	// 		continue
	// 	}
	// 	logger.Debugf("docker instance found at %s", host)
	// 	out = append(out, docker.NewPlatform(machine, cli))
	// }
	return out
}
