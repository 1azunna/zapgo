package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/1azunna/zapgo/internal/zapgo"
)

const (
	maxScanDurationInMins int = 30
)

type RunCommand struct {
	InitCommand
	File string `short:"f" long:"file" required:"true" description:"ZAP Automation framework config file. All files and folders in the current directory will be mounted to the /zap/wrk directory."`
}

// type RunOptions struct {
// 	Host     *url.URL
// 	Context  string
// 	Target   string
// 	Scripts  []struct{}
// 	Policies []string
// }

// var options RunOptions
var runCommand RunCommand

func (r *RunCommand) Execute(containerId string, baseUrl string, logger zapgo.Logger) {
	filename := filepath.Join("/zap/wrk/", r.File)
	logger.Info(fmt.Sprintf("Running automation framework with file %s", filename))
	// runOptions := &zapgo.RunOptions{
	// 	Host: baseUrl,
	// }
	// zapgo.Run(filename, runOptions)

	client := zapgo.ZapClient(baseUrl)
	plan, err := client.Automation().RunPlan(filename)
	if err != nil {
		panic(err)
	}
	for k, v := range plan {
		logger.Debug(fmt.Sprintf("%s:%s", k, v))
		if k == "code" {
			logger.Error(plan["message"].(string))
			r.TearDown(containerId, logger)
		}
	}
	planId := plan["planId"].(string)
	c := time.Tick(10 * time.Second)

	index := 0
	maxIndexSecs := 30 * 60
	maxIndex := maxIndexSecs / 10

	for range c {
		finished := false
		//Download the current contents of the URL and do something with it
		resp, err := client.Automation().PlanProgress(planId)
		if err != nil {
			panic(err)
		}
		// Check if the status is finished
		index = index + 1
		if index == maxIndex {
			logger.Error("Plan Timout Exceeded")
		}
		// TO-Do Implement check for plan errors
		for k := range resp {
			if k == "finished" {
				finished = true
				logger.Info("Automation plan complete")
				break
			}
		}
		if finished {
			break
		}

	}
	// for _, backoff := range retrySchedule {
	// 	status, err := client.Automation().PlanProgress(planId)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println(status)
	// 	time.Sleep(backoff)
	// }

	r.TearDown(containerId, logger)
}

func (r *RunCommand) TearDown(container string, logger zapgo.Logger) {
	zapgo.RemoveZapContainer(container, logger)
}

func init() {
	_, err := parser.AddCommand("run",
		"Run ZAP scan",
		"Run ZAP scan using the ZAP Automation framework. The automation file is required.",
		&runCommand)
	if err != nil {
		panic(err)
	}
}
