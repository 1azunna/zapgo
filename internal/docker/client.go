package docker

import (
	"github.com/1azunna/zapgo/internal/types"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type Docker struct {
	*types.Zapgo
}

// NewClient returns an object to communicate with the Docker Engine API.
func NewClient() *client.Client {

	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		logrus.Errorf("Could not create a docker client due to %v", err)
	}
	return client
}
