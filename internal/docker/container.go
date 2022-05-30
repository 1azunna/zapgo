package docker

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/1azunna/zapgo/internal/defaults"
	"github.com/1azunna/zapgo/internal/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
)

type containerClient interface {
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig,
		platform *specs.Platform, containerName string) (container.ContainerCreateCreatedBody, error)
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
}

func (c Docker) IfContainerExists(dockerClient containerClient, container string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), defaults.TimeoutInS)
	defer cancel()

	resp, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("name", container),
		),
	})
	if err != nil {
		logrus.Fatalf("Failed to list containers with name %s due to %v", container, err)
	}
	if len(resp) != 1 {
		return "", false
	}
	return resp[0].ID, true
}

func (c Docker) RemoveContainer(dockerClient containerClient, containerID string) {
	ctx, cancel := context.WithTimeout(context.Background(), defaults.TimeoutInS)
	defer cancel()

	err := dockerClient.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		logrus.Fatalf("Failed to remove container %s due to %v", containerID, err)
	}
	logrus.Debugf("Removed container %s", containerID)
}

func (c Docker) ContainerLogs(dockerClient containerClient, containerID string) {
	ctx, cancel := context.WithTimeout(context.Background(), defaults.TimeoutInS)
	defer cancel()

	reader, err := dockerClient.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		Timestamps: false,
		Follow:     false,
	})
	if err != nil {
		logrus.Fatal(err)
	}
	logs, err := ioutil.ReadAll(reader)
	if err != nil {
		logrus.Fatal(err)
	}
	err = os.WriteFile("zap.log", logs, 0600)
	if err != nil {
		logrus.Fatal(err)
	}
}

func (c Docker) RunZap(dockerClient containerClient) string {
	ctx, cancel := context.WithTimeout(context.Background(), defaults.TimeoutInS)
	defer cancel()

	startCommand := []string{"sh", "-c",
		fmt.Sprintf("zap-x.sh -daemon -port %s -host 0.0.0.0 -config api.disablekey=true -config api.addrs.addr.name=\".*\" -config api.addrs.addr.regex=true %s",
			c.ZapConfig.Port, strings.Join(c.ZapConfig.Options, " "))}
	dir := utils.CurrentDir()
	containerPort, err := nat.NewPort("tcp", c.ZapConfig.Port)
	if err != nil {
		logrus.Fatal(err)
	}
	config := &container.Config{
		Hostname:     c.ZapConfig.Hostname,
		ExposedPorts: nat.PortSet{containerPort: struct{}{}},
		Cmd:          startCommand,
		Image:        c.ZapConfig.Image,
	}
	host_config := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/zap/wrk", dir),
		},
		AutoRemove: true,
		PortBindings: nat.PortMap{
			containerPort: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: c.ZapConfig.Port,
				},
			},
		},
	}
	network_config := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			c.ZapConfig.Network: {
				NetworkID: c.ZapConfig.Network,
			},
		},
	}

	containerID, ifExists := c.IfContainerExists(dockerClient, c.ZapConfig.Container)
	if ifExists {
		logrus.Infof("The %s container already exists with ID %s", c.ZapConfig.Container, containerID)
		c.RemoveContainer(dockerClient, containerID)
	}

	logrus.Info(fmt.Sprintf("Creating new container: %s...", c.ZapConfig.Container))
	resp, err := dockerClient.ContainerCreate(ctx, config, host_config, network_config, nil, c.ZapConfig.Container)
	if err != nil {
		logrus.Fatalf("Failed to create container %s due to %v", c.ZapConfig.Container, err)
	}
	logrus.Debugf("Created the %s container with ID %s", c.ZapConfig.Container, resp.ID)

	// If the container is already running, Docker does not return an error response.
	if err := dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		logrus.Fatalf("Failed to start ZAP container with ID %s due to %v", resp.ID, err)
	}

	logrus.Debugf("Started ZAP container with ID %s", resp.ID)
	return resp.ID

}

//
//Newman Docker Container
//

func (c Docker) RunNewman(dockerClient containerClient) {
	ctx, cancel := context.WithTimeout(context.Background(), defaults.TimeoutInS)
	defer cancel()

	startCommand := []string{"run", c.PmConfig.Collection, "-e", c.PmConfig.Environment, "--reporter-cli-no-failures", "--reporter-cli-no-assertions", "--reporter-cli-no-console", "--insecure"}
	if c.PmConfig.Environment == "" {
		startCommand = []string{"run", c.PmConfig.Collection}
	}
	dir := utils.CurrentDir()

	config := &container.Config{
		Hostname: "newman",
		Env:      []string{fmt.Sprintf("http_proxy=%s:%s/", c.ZapConfig.Hostname, c.ZapConfig.Port), fmt.Sprintf("https_proxy=%s:%s/", c.ZapConfig.Hostname, c.ZapConfig.Port)},
		Cmd:      startCommand,
		Image:    c.PmConfig.Image,
	}
	host_config := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/etc/newman", dir),
		},
		AutoRemove: true,
	}
	network_config := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			c.ZapConfig.Network: {
				NetworkID: c.ZapConfig.Network,
			},
		},
	}

	containerID, ifExists := c.IfContainerExists(dockerClient, c.PmConfig.Container)
	if ifExists {
		logrus.Infof("The %s container already exists with ID %s", c.PmConfig.Container, containerID)
		c.RemoveContainer(dockerClient, containerID)
	}

	logrus.Infof("Creating new container: %s...", c.PmConfig.Container)
	resp, err := dockerClient.ContainerCreate(ctx, config, host_config, network_config, nil, c.PmConfig.Container)
	if err != nil {
		logrus.Errorf("Failed to create container %s due to %v", c.PmConfig.Container, err)
	} else {
		logrus.Debugf("Created the %s container with ID %s", c.PmConfig.Container, resp.ID)
	}

	// If the container is already running, Docker does not return an error response.
	if err := dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		logrus.Errorf("Failed to start container with ID %s due to %v", resp.ID, err)
	} else {
		logrus.Debugf("Started Newman container with ID %s", resp.ID)
	}
	//Output Newman Logs
	reader, err := dockerClient.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Timestamps: false,
		Follow:     true,
		Tail:       "40",
	})
	if err != nil {
		return
		// logrus.Fatal(err)
	}
	if _, err = io.Copy(os.Stdout, reader); err != nil && err != io.EOF {
		return // logrus.Fatal(err)
	}
}
