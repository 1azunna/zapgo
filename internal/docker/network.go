package docker

import (
	"context"

	"github.com/1azunna/zapgo/internal/defaults"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type networkClient interface {
	NetworkInspect(ctx context.Context, networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error)
	NetworkCreate(ctx context.Context, name string, options types.NetworkCreate) (types.NetworkCreateResponse, error)
	NetworkRemove(ctx context.Context, networkID string) error
}

func (c Docker) IfZapNetworkExists(dockerClient networkClient) bool {
	ctx, cancel := context.WithTimeout(context.Background(), defaults.TimeoutInS)
	defer cancel()

	_, err := dockerClient.NetworkInspect(ctx, c.ZapConfig.Network, types.NetworkInspectOptions{})
	if err != nil {
		if client.IsErrNotFound(err) {
			return false
		}
		// Unexpected error while inspecting docker networks, we want to crash the app.
		logrus.Fatalf("Failed to inspect docker network %s due to %v", c.ZapConfig.Network, err)
	}
	return true
}

func createZapNetwork(dockerClient networkClient, network string) {
	ctx, cancel := context.WithTimeout(context.Background(), defaults.TimeoutInS)
	defer cancel()

	logrus.Infof("Creating network: %s...", network)
	resp, err := dockerClient.NetworkCreate(ctx, network, types.NetworkCreate{})
	if err != nil {
		logrus.Fatalf("Failed to create network %s due to %v", network, err)
	}
	logrus.Debugf("Created network %s with ID %s", network, resp.ID)
}

//Setup ZAP Network
func (c Docker) SetupZapNetwork(dockerClient networkClient) {

	if c.IfZapNetworkExists(dockerClient) {
		logrus.Infof("The network %s already exists", c.ZapConfig.Network)
		return
	}
	createZapNetwork(dockerClient, c.ZapConfig.Network)
}

//Remove ZAP Network
func (c Docker) RemoveZapNetwork(dockerClient networkClient) {
	ctx, cancel := context.WithTimeout(context.Background(), defaults.TimeoutInS)
	defer cancel()

	logrus.Infof("Removing network: %s...", c.ZapConfig.Network)
	err := dockerClient.NetworkRemove(ctx, c.ZapConfig.Network)
	if err != nil {
		logrus.Fatalf("Failed to remove network %s due to %v", c.ZapConfig.Network, err)
	}
	logrus.Debugf("Removed network %s", c.ZapConfig.Network)
}
