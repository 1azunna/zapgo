package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/1azunna/zapgo/internal/zapgo"
	"github.com/sirupsen/logrus"
)

type InitCommand struct {
	Release string   `long:"release" choice:"stable" choice:"weekly" default:"weekly" description:"The image release version"`
	Port    string   `long:"port" default:"8080" description:"Initialize ZAP with a custom port. Default is 8080"`
	Pull    bool     `short:"p" long:"pull" description:"Pull the latest ZAP image from dockerhub"`
	Configs []string `long:"extraConfig" description:"Additional ZAP configurations to use when initializing ZAP"`
}

var retrySchedule = []time.Duration{
	5 * time.Second,
	10 * time.Second,
	15 * time.Second,
	30 * time.Second,
}

var initCommand InitCommand

func (i *InitCommand) Execute(logger zapgo.Logger) (string, string) {

	logger.Info(fmt.Sprintf("Using the %v release", i.Release))
	imageStr := fmt.Sprintf("owasp/zap2docker-%v", i.Release)
	// Pull ZAP image
	if i.Pull {
		zapgo.PullZapImage(imageStr, logger)
	} else if !zapgo.ZapImageExists(imageStr) {
		zapgo.PullZapImage(imageStr, logger)
	}
	// Initialize ZAP Network
	zapgo.SetupZapNetwork(logger)
	// Create ZAP Container
	containerID := zapgo.CreateZapContainer(imageStr, i.Port, i.Configs, logger)
	// Start the ZAP Container
	zapgo.StartZapContainer(containerID, logger)
	// Return the zap host url
	base := fmt.Sprintf("http://localhost:%s", i.Port)
	i.HealthCheck(logger)
	return containerID, base
}

func (i *InitCommand) HealthCheck(logger zapgo.Logger) {
	// Wait 10seconds before checking for liveness
	time.Sleep(10 * time.Second)
	for _, backoff := range retrySchedule {
		_, err := http.Get(fmt.Sprintf("http://localhost:%s", i.Port))
		if err == nil {
			logger.Info("ZAP container initialized successfully!")
			break
		}
		logger.Warn("ZAP is not ready")
		logger.Warn(fmt.Sprintf("Retrying in %v", backoff))
		time.Sleep(backoff)
	}
	resp, _ := http.Get(fmt.Sprintf("http://localhost:%s", i.Port))
	if resp == nil {
		logger.Error("Could not reach ZAP container")
		os.Exit(1)
	}
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
