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

var parser = flags.NewParser(&options, flags.Default)

var introtext string
var logLevel string

func main() {
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

	_, err := parser.Parse()
	if err != nil {
		if len(os.Args) == 1 {
			fmt.Println(introtext)
		} else {
			fmt.Printf("See zapgo %v --help for usage details.\n", os.Args[1])
		}
		os.Exit(1)
	}

	switch parser.Active.Name {
	case "init":
		initCommand.Execute(logger)
		// Check if the server is reachable
		initCommand.HealthCheck(logger)

	default:
		os.Exit(1)
	}

}
