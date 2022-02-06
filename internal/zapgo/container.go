package zapgo

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
)

type ZapOptions struct {
	Image     string
	Hostname  string
	Container string
	Network   string
	Port      string
	Options   []string
}

type NewmanOptions struct {
	NewmanImage     string
	NewmanContainer string
	Collection      string
	Environment     string
}

type Zapgo struct {
	*ZapOptions
	*NewmanOptions
}

func currentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
	}
	return dir
}

type containerClient interface {
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig,
		platform *specs.Platform, containerName string) (container.ContainerCreateCreatedBody, error)
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
}

func (z *Zapgo) IfContainerExists(dockerClient containerClient, container string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutInS)
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

func (z *Zapgo) RemoveContainer(dockerClient containerClient, containerID string) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutInS)
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

func (z *Zapgo) ContainerLogs(dockerClient containerClient, containerID string) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutInS)
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

func (z *Zapgo) RunZap(dockerClient containerClient) string {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutInS)
	defer cancel()

	startCommand := []string{"sh", "-c",
		fmt.Sprintf("zap-x.sh -daemon -port %s -host 0.0.0.0 -config api.disablekey=true -config api.addrs.addr.name=\".*\" -config api.addrs.addr.regex=true %s",
			z.Port, strings.Join(z.Options, " "))}
	dir := currentDir()
	containerPort, err := nat.NewPort("tcp", z.Port)
	if err != nil {
		logrus.Fatal(err)
	}
	config := &container.Config{
		Hostname:     z.Hostname,
		ExposedPorts: nat.PortSet{containerPort: struct{}{}},
		Cmd:          startCommand,
		Image:        z.Image,
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
					HostPort: z.Port,
				},
			},
		},
	}
	network_config := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			z.Network: {
				NetworkID: z.Network,
			},
		},
	}

	containerID, ifExists := z.IfContainerExists(dockerClient, z.Container)
	if ifExists {
		logrus.Infof("The %s container already exists with ID %s", z.Container, containerID)
		z.RemoveContainer(dockerClient, containerID)
	}

	logrus.Info(fmt.Sprintf("Creating new container: %s...", z.Container))
	resp, err := dockerClient.ContainerCreate(ctx, config, host_config, network_config, nil, z.Container)
	if err != nil {
		logrus.Fatalf("Failed to create container %s due to %v", z.Container, err)
	}
	logrus.Debugf("Created the %s container with ID %s", z.Container, resp.ID)

	// If the container is already running, Docker does not return an error response.
	if err := dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		logrus.Fatalf("Failed to start ZAP container with ID %s due to %v", resp.ID, err)
	}

	logrus.Debugf("Started ZAP container with ID %s", resp.ID)
	return resp.ID

}

//
//Newman Docker
//

func (z *Zapgo) RunNewman(dockerClient containerClient) {
	var startCommand []string
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutInS)
	defer cancel()

	startCommand = []string{"run", z.Collection, "-e", z.Environment, "--reporter-cli-no-failures", "--reporter-cli-no-assertions", "--reporter-cli-no-console", "--insecure"}
	if z.Environment == "" {
		startCommand = []string{"run", z.Collection}
	}
	dir := currentDir()

	config := &container.Config{
		Hostname: "newman",
		Env:      []string{fmt.Sprintf("http_proxy=%s:%s/", z.Hostname, z.Port), fmt.Sprintf("https_proxy=%s:%s/", z.Hostname, z.Port)},
		Cmd:      startCommand,
		Image:    z.NewmanImage,
	}
	host_config := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/etc/newman", dir),
		},
		AutoRemove: true,
	}
	network_config := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			z.Network: {
				NetworkID: z.Network,
			},
		},
	}

	containerID, ifExists := z.IfContainerExists(dockerClient, z.NewmanContainer)
	if ifExists {
		logrus.Infof("The %s container already exists with ID %s", z.NewmanContainer, containerID)
		z.RemoveContainer(dockerClient, containerID)
	}

	logrus.Infof("Creating new container: %s...", z.NewmanContainer)
	resp, err := dockerClient.ContainerCreate(ctx, config, host_config, network_config, nil, z.NewmanContainer)
	if err != nil {
		logrus.Fatalf("Failed to create container %s due to %v", z.NewmanContainer, err)
	}

	logrus.Debugf("Created the %s container with ID %s", z.NewmanContainer, resp.ID)

	// If the container is already running, Docker does not return an error response.
	if err := dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		logrus.Fatalf("Failed to start container with ID %s due to %v", resp.ID, err)
	}
	logrus.Debugf("Started Newman container with ID %s", resp.ID)
	//Output Newman Logs
	reader, err := dockerClient.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Timestamps: false,
		Follow:     true,
		Tail:       "40",
	})
	if err != nil {
		logrus.Fatal(err)
	}
	if _, err = io.Copy(os.Stdout, reader); err != nil && err != io.EOF {
		logrus.Fatal(err)
	}
}
