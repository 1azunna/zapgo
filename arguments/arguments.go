package arguments

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

type cmdArgs struct {
	// Example of verbosity with level
	Verbose []bool   `short:"v" long:"verbose" description:"Show verbose output"`
	Release string   `long:"release" choice:"stable" choice:"weekly" choice:"live" choice:"bare" default:"stable" description:"The docker image tag to use"`
	Port    *int     `long:"port" default:"8080" description:"Initialize ZAP with a custom port."`
	Pull    bool     `short:"p" long:"pull" description:"Pull the latest ZAP image from dockerhub"`
	Configs []string `long:"opts" description:"Additional ZAP command line options to use when initializing ZAP"`
}

var Options cmdArgs
var Parser = flags.NewParser(&Options, flags.Default)

func CheckArgs() {
	introtext := `ZapGo is a command line utility for dynamic security testing based on the OWASP ZAP Project.
	See zapgo --help for usage details.`

	if _, err := Parser.Parse(); err != nil {
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
}
