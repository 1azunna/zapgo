package cmd

import (
	args "github.com/1azunna/zapgo/arguments"
	"github.com/1azunna/zapgo/internal/docker"
	"github.com/1azunna/zapgo/internal/utils"
	"github.com/sirupsen/logrus"
)

type InitOpts struct {
	NetworkOnly bool `short:"n" long:"networkOnly" description:"Create the zapgo-network without initializing the ZAP container."`
}

var initCmd InitOpts

func (i *InitOpts) Execute(zapgo *docker.Docker) string {

	if i.NetworkOnly {
		zapgo.SetupZapNetwork(client)
	} else {
		// Pull ZAP image
		if args.Options.Pull {
			zapgo.PullImage(client, zapgo.ZapConfig.Image)
		} else if !zapgo.ImageExists(client, zapgo.ZapConfig.Image) {
			zapgo.PullImage(client, zapgo.ZapConfig.Image)
		}
		// Initialize ZAP Network
		zapgo.SetupZapNetwork(client)
		// Create ZAP Container
		containerID := zapgo.RunZap(client)

		utils.HealthCheck(baseURL)
		return containerID
	}
	return ""
}

func init() {
	_, err := args.Parser.AddCommand("init",
		"Initialize ZAP",
		"The init command pulls the zap image from the docker registry.",
		&initCmd)
	if err != nil {
		logrus.Fatal(err)
	}
}
