package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	args "github.com/1azunna/zapgo/arguments"
	"github.com/1azunna/zapgo/internal/docker"
	"github.com/1azunna/zapgo/internal/types"
	"github.com/sirupsen/logrus"
)

// Create docker client
var client = docker.NewClient()

var baseURL string

func Execute() {

	args.CheckArgs()
	// Create logger
	if len(args.Options.Verbose) > 0 {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	imageStr := fmt.Sprintf("owasp/zap2docker-%v", args.Options.Release)
	logrus.Debugf("Using the %v release %s", args.Options.Release, imageStr)

	if *args.Options.Port < 1 || *args.Options.Port > 65535 {
		logrus.Fatalf("Using and invalid port range. Please specify a value between 1 and 65535")
	}

	zapOpts := types.Zaproxy{
		Image:     imageStr,
		Hostname:  "zap",
		Container: "zapgo-container",
		Network:   "zapgo-network",
		Port:      strconv.Itoa(*args.Options.Port),
		Options:   args.Options.Configs,
	}

	pmOpts := types.Newman{
		Image:       "postman/newman",
		Container:   "zapgo-newman",
		Collection:  runCmd.Collection,
		Environment: runCmd.Environment,
	}

	zapgo := types.Zapgo{
		ZapConfig: &zapOpts,
		PmConfig:  &pmOpts,
	}
	// Zapgo Docker Runtime
	zapgoDocker := &docker.Docker{
		&zapgo,
	}
	// Return the zap host url
	baseURL = fmt.Sprintf("http://localhost:%s", zapOpts.Port)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-sigc
		logrus.Warnln(<-sigc, "signal received")
		cleanCmd.Execute(zapgoDocker)
	}()

	switch args.Parser.Active.Name {
	case "init":
		initCmd.Execute(zapgoDocker)
	case "run":
		runCmd.Execute(zapgoDocker)
	case "clean":
		cleanCmd.Execute(zapgoDocker)
	default:
		os.Exit(1)
	}

}
