package cmd

import (
	args "github.com/1azunna/zapgo/arguments"
	"github.com/1azunna/zapgo/internal/docker"
	"github.com/sirupsen/logrus"
)

type CleanOpts struct{}

var cleanCmd CleanOpts

func cleanContainers(zapgo *docker.Docker) {

	zapContainerId, ifZapExists := zapgo.IfContainerExists(client, zapgo.ZapConfig.Container)
	if ifZapExists {
		logrus.Infof("Removing %s container....", zapgo.ZapConfig.Container)
		zapgo.RemoveContainer(client, zapContainerId)
	}
	newmanContainerId, ifNewmanExists := zapgo.IfContainerExists(client, zapgo.PmConfig.Container)
	if ifNewmanExists {
		logrus.Infof("Removing %s container....", zapgo.PmConfig.Container)
		zapgo.RemoveContainer(client, newmanContainerId)
	}
}

func (c *CleanOpts) Execute(zapgo *docker.Docker) {

	cleanContainers(zapgo)
	if zapgo.IfZapNetworkExists(client) {
		zapgo.RemoveZapNetwork(client)
	}
}

func init() {
	_, err := args.Parser.AddCommand("clean",
		"Clean Zapgo",
		"The Clean command removes Zapgo containers and networks.",
		&cleanCmd)
	if err != nil {
		logrus.Fatal(err)
	}
}
