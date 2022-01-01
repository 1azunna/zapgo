package main

import (
	"fmt"
	"os"

	"github.com/1azunna/zapgo/internal/zapgo"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	// Example of verbosity with level
	Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
}

var options Options
var introtext string

var parser = flags.NewParser(&options, flags.Default)

func main() {
	var logLevel string
	// var baseUrl *url.URL

	introtext = `ZapGo is a command line utility for dynamic security testing based on the OWASP ZAP Project.
	See zapgo --help for usage details.`

	// Create logger
	if len(options.Verbose) > 0 {
		logLevel = "debug"
	} else {
		logLevel = "info"
	}

	logger := zapgo.NewLogger(zapgo.LogConfig{
		Format:  "logfmt",
		Level:   logLevel,
		NoColor: false,
	})
	zapgo.SetStandardLogger(logger)

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

	switch parser.Active.Name {
	case "init":
		initCommand.Execute(logger)
	case "run":
		containerId, baseUrl := initCommand.Execute(logger)
		runCommand.Execute(containerId, baseUrl, logger)
	default:
		os.Exit(1)
	}

}
