package main

import (
	"github.com/1azunna/zapgo/internal/zapgo"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type InitOptions struct {
	NetworkOnly bool `short:"n" long:"networkOnly" description:"Create the zapgo-network without initializing the ZAP container."`
}

var initCommand InitOptions

func (i *InitOptions) Execute(zap *zapgo.Zapgo, client *client.Client) string {

	if i.NetworkOnly {
		zap.SetupZapNetwork(client)
	} else {
		// Pull ZAP image
		if options.Pull {
			zap.PullImage(client, zap.Image)
		} else if !zap.ImageExists(client, zap.Image) {
			zap.PullImage(client, zap.Image)
		}
		// Initialize ZAP Network
		zap.SetupZapNetwork(client)
		// Create ZAP Container
		containerID := zap.RunZap(client)

		zap.HealthCheck(BaseURL)
		return containerID
	}
	return ""
}

func init() {
	_, err := parser.AddCommand("init",
		"Initialize ZAP",
		"The init command pulls the zap image from the docker registry.",
		&initCommand)
	if err != nil {
		logrus.Fatal(err)
	}
}
