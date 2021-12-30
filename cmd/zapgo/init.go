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
	Release string `long:"release" choice:"stable" choice:"weekly" default:"weekly" subcommands-optional:"init" description:"The image release version"`
	Port    string `long:"port" default:"8080" description:"Initialize ZAP with a custom port. Default is 8080"`
	Pull    bool   `short:"p" long:"pull" description:"Pull the latest ZAP image from dockerhub"`
}

var retrySchedule = []time.Duration{
	5 * time.Second,
	10 * time.Second,
	15 * time.Second,
	30 * time.Second,
}

var initCommand InitCommand

func (i *InitCommand) Execute(logger zapgo.Logger) {

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
	containerID := zapgo.CreateZapContainer(imageStr, i.Port, logger)
	// Start the ZAP Container
	zapgo.StartZapContainer(containerID, logger)
}

func (i *InitCommand) HealthCheck(logger zapgo.Logger) {
	time.Sleep(10 * time.Second)
	for _, backoff := range retrySchedule {
		_, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s", initCommand.Port))
		if err == nil {
			logger.Info("ZAP container initialized successfully!")
			break
		}
		logger.Warn("ZAP is not ready")
		logger.Warn(fmt.Sprintf("Retrying in %v", backoff))
		time.Sleep(backoff)
	}
	resp, _ := http.Get(fmt.Sprintf("http://127.0.0.1:%s", initCommand.Port))
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
