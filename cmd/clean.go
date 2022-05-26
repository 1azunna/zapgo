package cmd

import (
	"github.com/1azunna/zapgo/internal/zapgo"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type CleanOptions struct{}

var cleanCommand CleanOptions

func cleanContainers(zap *zapgo.Zapgo, client *client.Client) {

	zapContainerId, ifZapExists := zap.IfContainerExists(client, zap.Container)
	if ifZapExists {
		logrus.Infof("Removing %s container....", zap.Container)
		zap.RemoveContainer(client, zapContainerId)
	}
	newmanContainerId, ifNewmanExists := zap.IfContainerExists(client, zap.NewmanContainer)
	if ifNewmanExists {
		logrus.Infof("Removing %s container....", zap.NewmanContainer)
		zap.RemoveContainer(client, newmanContainerId)
	}
}

func (c *CleanOptions) Execute(zap *zapgo.Zapgo, client *client.Client) {

	cleanContainers(zap, client)
	if zap.IfZapNetworkExists(client) {
		zap.RemoveZapNetwork(client)
	}
}

func init() {
	_, err := parser.AddCommand("clean",
		"Clean Zapgo",
		"The Clean command removes Zapgo containers and networks.",
		&cleanCommand)
	if err != nil {
		logrus.Fatal(err)
	}
}
