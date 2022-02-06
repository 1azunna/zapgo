package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/1azunna/zapgo/internal/zapgo"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

type Options struct {
	// Example of verbosity with level
	Verbose []bool   `short:"v" long:"verbose" description:"Show verbose output"`
	Release string   `long:"release" choice:"stable" choice:"weekly" default:"weekly" description:"The image release version"`
	Port    *int     `long:"port" default:"8080" description:"Initialize ZAP with a custom port."`
	Pull    bool     `short:"p" long:"pull" description:"Pull the latest ZAP image from dockerhub"`
	Configs []string `long:"opts" description:"Additional ZAP command line options to use when initializing ZAP"`
}

var options Options
var introtext string
var BaseURL string

var parser = flags.NewParser(&options, flags.Default)

func main() {

	introtext = `ZapGo is a command line utility for dynamic security testing based on the OWASP ZAP Project.
	See zapgo --help for usage details.`

	if _, err := parser.Parse(); err != nil {
		if len(os.Args) == 1 {
			fmt.Println(introtext)
		}
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}

	// Create logger
	if len(options.Verbose) > 0 {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	imageStr := fmt.Sprintf("owasp/zap2docker-%v", options.Release)
	logrus.Debugf("Using the %v release %s", options.Release, imageStr)

	if *options.Port < 1 || *options.Port > 65535 {
		logrus.Fatalf("Using and invalid port range. Please specify a value between 1 and 65535")
	}

	zap := &zapgo.Zapgo{
		&zapgo.ZapOptions{
			Image:     imageStr,
			Hostname:  "zap",
			Container: "zapgo-container",
			Network:   "zapgo-network",
			Port:      strconv.Itoa(*options.Port),
			Options:   options.Configs,
		},
		&zapgo.NewmanOptions{
			NewmanImage:     "postman/newman",
			NewmanContainer: "zapgo-newman",
			Collection:      runCommand.Collection,
			Environment:     runCommand.Environment,
		},
	}
	// Create docker client
	client := zapgo.NewClient()
	// Return the zap host url
	BaseURL = fmt.Sprintf("http://localhost:%s", zap.Port)

	switch parser.Active.Name {
	case "init":
		initCommand.Execute(zap, client)
	case "run":
		runCommand.Execute(zap, client)
	case "clean":
		cleanCommand.Execute(zap, client)
	default:
		os.Exit(1)
	}

}
