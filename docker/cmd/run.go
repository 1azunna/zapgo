package cmd

import (
	"time"

	args "github.com/1azunna/zapgo/arguments"
	"github.com/1azunna/zapgo/internal/docker"
	"github.com/1azunna/zapgo/internal/types"
	"github.com/1azunna/zapgo/internal/utils"
	"github.com/1azunna/zapgo/internal/zaproxy"
	"github.com/sirupsen/logrus"
)

type RunOpts struct {
	File        string `long:"file" required:"true" description:"ZAP Automation framework config file. Automation file must be placed within the current working directory."`
	Collection  string `long:"collection" description:"Postman collection file or url to run."`
	Environment string `long:"environment" description:"Postman environment file or url to use with postman collection"`
	Policy      string `long:"policy" description:"Import custom zap scan policy. Policy file must be placed within the current working directory."`
	Clean       bool   `short:"c" long:"clean" description:"Remove any existing zapgo containers and initialize ZAP."`
	Confidence  string `long:"confidence" default:"Medium" choice:"Low" choice:"Medium" choice:"High" choice:"Confirmed" description:"Display alerts with confidence filter set to either Low, Medium, High or Confirmed."`
	Risk        string `long:"risk" default:"Low" choice:"Low" choice:"Medium" choice:"High" choice:"Informational" description:"Display alerts with risk filter set to either Informational, Low, Medium, High."`
	Fail        string `long:"fail" choice:"Low" choice:"Medium" choice:"High" description:"Set exit status to fail on a certain risk level. Allowed Risk levels are Low|Medium|High."`
	Display     string `long:"display" choice:"Sites" choice:"Contexts" choice:"All" default:"All" description:"Set display output format for alerts found."`
}

var runCmd RunOpts

func newman(zapgo *docker.Docker) {
	if !zapgo.ImageExists(client, zapgo.PmConfig.Image) {
		zapgo.PullImage(client, zapgo.PmConfig.Image)
	}
	zapgo.RunNewman(client)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if _, ifExists := zapgo.IfContainerExists(client, zapgo.PmConfig.Container); !ifExists {
			break
		}
	}
}

func (r *RunOpts) Execute(zapgo *docker.Docker) {

	var containerId string

	// TODO: Support user specified paths.
	path := utils.CurrentDir()
	filters := types.AlertFilters{
		Confidence: r.Confidence,
		Risk:       r.Risk,
	}

	if r.Clean {
		cleanContainers(zapgo)
		id := initCmd.Execute(zapgo)
		containerId = id
	} else {
		id, ifExists := zapgo.IfContainerExists(client, zapgo.ZapConfig.Container)
		containerId = id
		if !ifExists {
			logrus.Fatal("ZAP is not running. Run \"zapgo init\" OR add the \"--clean\" flag to \"zapgo run\"")
		}
	}

	//Run postman collections if available
	if r.Collection != "" {
		newman(zapgo)
	}

	// setup zap client
	zap := zaproxy.ZapClient(baseURL)

	// import any specified policies.
	zaproxy.ImportPolicy(r.Policy, path, zap)

	// Run automation plan
	planId := zaproxy.RunPlan(r.File, path, zap)

	// Get plan status
	zaproxy.GetPlanStatus(planId, zap)

	// Get scan results
	zaproxy.GetScanResults(r.File, planId, filters, r.Display, r.Fail, zap)

	//Print out zap log
	zapgo.ContainerLogs(client, containerId)

	// Quit Zap
	zapgo.RemoveContainer(client, containerId)
	// zap.RemoveZapNetwork(client)

}

func init() {
	_, err := args.Parser.AddCommand("run",
		"Run ZAP scan",
		"Run ZAP scan using the ZAP Automation framework. The automation file is required.",
		&runCmd)
	if err != nil {
		panic(err)
	}
}
