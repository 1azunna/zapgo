package zapgo

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
)

type networkCreator interface {
	NetworkInspect(ctx context.Context, networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error)
	NetworkCreate(ctx context.Context, name string, options types.NetworkCreate) (types.NetworkCreateResponse, error)
	NetworkRemove(ctx context.Context, networkID string) error
}

func ifZapNetworkExists(dockerClient networkCreator) bool {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutInS)
	defer cancel()

	_, err := dockerClient.NetworkInspect(ctx, NetworkName, types.NetworkInspectOptions{})
	if err != nil {
		if client.IsErrNotFound(err) {
			return false
		}
		// Unexpected error while inspecting docker networks, we want to crash the app.
		logrus.Fatalf("Failed to inspect docker network %s due to %v", NetworkName, err)
	}
	return true
}

func createZapNetwork(dockerClient networkCreator, logger Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutInS)
	defer cancel()

	logger.Info(fmt.Sprintf("Creating network: %s...", NetworkName))
	resp, err := dockerClient.NetworkCreate(ctx, NetworkName, types.NetworkCreate{})
	if err != nil {
		logrus.Fatalf("Failed to create network %s due to %v", NetworkName, err)
	}
	logger.Info(fmt.Sprintf("Created network %s with ID %s", NetworkName, resp.ID))
}

//Setup ZAP Network
func SetupZapNetwork(logger Logger) {

	dockerClient := NewClient()
	if ifZapNetworkExists(dockerClient) {
		logger.Info(fmt.Sprintf("The network %s already exists", NetworkName))
		return
	}
	createZapNetwork(dockerClient, logger)
}

func RemoveZapNetwork(logger Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutInS)
	defer cancel()

	dockerClient := NewClient()
	logger.Info(fmt.Sprintf("Removing network: %s...", NetworkName))
	err := dockerClient.NetworkRemove(ctx, NetworkName)
	if err != nil {
		logrus.Fatalf("Failed to remove network %s due to %v", NetworkName, err)
	}
	logger.Info(fmt.Sprintf("Removed network %s", NetworkName))
}

func ifZapContainerExists() (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutInS)
	defer cancel()
	dockerClient := NewClient()

	resp, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("name", ContainerName),
		),
	})
	if err != nil {
		logrus.Fatalf("Failed to list containers with name %s due to %v", ContainerName, err)
	}
	if len(resp) != 1 {
		return "", false
	}
	return resp[0].ID, true
}

func CreateZapContainer(imageName string, zapPort string, logger Logger) string {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutInS)
	defer cancel()

	dockerClient := NewClient()
	startCommand := []string{"sh", "-c", fmt.Sprintf("zap-x.sh -daemon -port %s -host 0.0.0.0 -config api.disablekey=true -config api.addrs.addr.name=\".*\" -config api.addrs.addr.regex=true", zapPort)}
	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
	}
	containerPort, err := nat.NewPort("tcp", zapPort)
	if err != nil {
		logrus.Fatal(err)
	}
	config := &container.Config{
		Hostname:     "zap",
		ExposedPorts: nat.PortSet{containerPort: struct{}{}},
		Cmd:          startCommand,
		Image:        imageName,
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
					HostPort: zapPort,
				},
			},
		},
	}
	network_config := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			NetworkName: {
				NetworkID: NetworkName,
			},
		},
	}

	containerID, ifExists := ifZapContainerExists()
	if ifExists {
		logger.Info(fmt.Sprintf("The %s container already exists with ID %s", ContainerName, containerID))
		removeZapContainer(containerID, logger)
	}

	logger.Info(fmt.Sprintf("Creating new container: %s...", ContainerName))
	resp, err := dockerClient.ContainerCreate(ctx, config, host_config, network_config, nil, ContainerName)
	if err != nil {
		logrus.Fatalf("Failed to create container %s due to %v", ContainerName, err)
	}

	logger.Info(fmt.Sprintf("Created the %s container with ID %s", ContainerName, resp.ID))
	return resp.ID

}

func removeZapContainer(containerID string, logger Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutInS)
	defer cancel()

	dockerClient := NewClient()
	err := dockerClient.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		logrus.Fatalf("Failed to remove container %s due to %v", containerID, err)
	}
	logger.Info(fmt.Sprintf("Removed container %s", containerID))
}

func StartZapContainer(containerID string, logger Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutInS)
	defer cancel()

	dockerClient := NewClient()
	// If the container is already running, Docker does not return an error response.
	if err := dockerClient.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		logrus.Fatalf("Failed to start container with ID %s due to %v", containerID, err)
	}
	logger.Info(fmt.Sprintf("Started ZAP container with ID %s", containerID))
}
