package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/1azunna/zapgo/internal/zapgo"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

const (
	maxScanDurationInMins int = 30
)

type RunCommand struct {
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

var runCommand RunCommand

type RiskCount struct {
	High          []int
	Medium        []int
	Low           []int
	Informational []int
}

func Newman(zap *zapgo.Zapgo, client *client.Client) {
	if !zap.ImageExists(client, zap.NewmanImage) {
		zap.PullImage(client, zap.NewmanImage)
	}
	zap.RunNewman(client)
}

func countSum(arr []int) int {
	res := 0
	for i := 0; i < len(arr); i++ {
		res += arr[i]
	}
	return res
}

func (r *RunCommand) Execute(zap *zapgo.Zapgo, client *client.Client) {

	var containerId string
	// setup zap client
	ZapClient := ZapClient(BaseURL)

	if r.Clean {
		cleanContainers(zap, client)
		id := initCommand.Execute(zap, client)
		containerId = id
	} else {
		id, ifExists := zap.IfContainerExists(client, zap.Container)
		containerId = id
		if !ifExists {
			logrus.Fatal("ZAP is not running. Run \"zapgo init\" OR add the \"--clean\" flag to \"zapgo run\"")
		}
	}
	// Check for automation file
	wd, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(wd, r.File)); errors.Is(err, os.ErrNotExist) {
		logrus.Fatalf("The specified file %s does not exist.", filepath.Join(wd, r.File))
	}
	filename := filepath.Join("/zap/wrk/", r.File)
	//Check for policy file
	if _, err := os.Stat(filepath.Join(wd, r.Policy)); len(r.Policy) > 0 && errors.Is(err, os.ErrNotExist) {
		logrus.Fatalf("The specified file %s does not exist.", filepath.Join(wd, r.Policy))
	} else if len(r.Policy) > 0 && err == nil {
		policyfile := filepath.Join("/zap/wrk/", r.Policy)
		resp, err := ZapClient.Ascan().ImportScanPolicy(policyfile)
		if err != nil {
			panic(err)
		}
		// fmt.Println(resp)
		logrus.Debugf("importing scan policy %s : %s", filepath.Join(wd, r.Policy), resp["Result"])
	}
	// Useful zap scan Options
	resp, err := ZapClient.Pscan().SetScanOnlyInScope("true")
	if err != nil {
		panic(err)
	}
	logrus.Debugf("Setting Pscan option SetScanOnlyInScope: %s", resp["Result"])

	logrus.Infof("Running automation framework with file %s", filename)

	yamlData, err := ioutil.ReadFile(r.File)
	if err != nil {
		panic(fmt.Sprintf("Could not parse yaml file %s \n%s", r.File, err))
	}
	// Will use this later to print results.
	zapgo.PrintContexts(yamlData)

	//Run postman collections if available
	if zap.Collection != "" {
		Newman(zap, client)
		for range time.Tick(5 * time.Second) {
			if _, ifExists := zap.IfContainerExists(client, zap.NewmanContainer); !ifExists {
				break
			}
		}
	}

	plan, err := ZapClient.Automation().RunPlan(filename)
	if err != nil {
		panic(err)
	}
	for k, v := range plan {
		logrus.Debug(fmt.Sprintf("%s:%s", k, v))
		if k == "code" {
			logrus.Error(plan["message"].(string))
			zap.RemoveContainer(client, containerId)
		}
	}
	planId := plan["planId"].(string)
	c := time.Tick(10 * time.Second)
	index := 0
	maxIndexSecs := maxScanDurationInMins * 60
	maxIndex := maxIndexSecs / 10

	for range c {
		finished := false
		//Get the plan progress
		resp, err := ZapClient.Automation().PlanProgress(planId)
		if err != nil {
			panic(err)
		}
		// Check if the status is finished
		index = index + 1
		if index == maxIndex {
			logrus.Error("Plan Timeout Exceeded")
			os.Exit(1)
		}
		// TO-Do Implement check for plan errors
		for k := range resp {
			if k == "finished" {
				finished = true
				logrus.Info("Automation plan complete")
				break
			}
		}
		if finished {
			break
		}

	}

	//Get contexts
	var alertsList [][]string
	var totalCount RiskCount
	contexts := zapgo.GetContexts(yamlData)
	filters := zapgo.Filters{
		Confidence: r.Confidence,
		Risk:       r.Risk,
	}
	for _, v := range contexts {
		for _, url := range v.Urls {
			zapAlerts, err := ZapClient.Alert().AlertsByRisk(url, "true")
			if err != nil {
				panic(err)
			}
			// fmt.Println(zapAlerts)
			alerts := zapgo.GetAlerts(zapAlerts, filters)
			for i := 0; i < len(alerts); i++ {
				alertsList = append(alertsList, alerts[i])
			}
			zapAlertsCount, err := ZapClient.Alert().AlertCountsByRisk(url, "true")
			if err != nil {
				panic(err)
			}
			count := zapgo.GetAlertsCount(zapAlertsCount)
			totalCount.High = append(totalCount.High, count.High)
			totalCount.Medium = append(totalCount.Medium, count.Medium)
			totalCount.Low = append(totalCount.Low, count.Low)
			totalCount.Informational = append(totalCount.Informational, count.Informational)

			if r.Display == "Sites" {
				fmt.Printf("\n Alerts discovered on Site: %s \n", url)
				zapgo.PrintAlerts(alertsList)
			}
		}
		if r.Display == "Contexts" {
			fmt.Printf("\n Alerts discovered on Context: %s \n", v.Name)
			zapgo.PrintAlerts(alertsList)
		}
	}

	if r.Display == "All" {
		zapgo.PrintAlerts(alertsList)
	}
	//Print out zap log
	zap.ContainerLogs(client, containerId)

	// Quit Zap
	zap.RemoveContainer(client, containerId)
	// zap.RemoveZapNetwork(client)

	totalAlerts := countSum(totalCount.High) + countSum(totalCount.Medium) + countSum(totalCount.Low) + countSum(totalCount.Informational)
	logrus.Infof("A total of %s alerts were found. See zap.log for zap output.", strconv.Itoa(totalAlerts))
	// Set exit code 1 if fail is specified
	r.setExitCode(totalCount)
}

func (r *RunCommand) setExitCode(riskcount RiskCount) {
	v := reflect.ValueOf(riskcount)
	typeV := v.Type()
	var fail bool
	for i := 0; i < v.NumField(); i++ {
		arr := v.Field(i).Interface().([]int)
		if typeV.Field(i).Name == r.Fail {
			if countSum(arr) > 0 {
				fail = true
			} else {
				for x := i; x < v.NumField(); x++ {
					arr := v.Field(x).Interface().([]int)
					if countSum(arr) > 0 {
						fail = true
					}
				}
			}
		}
	}
	if fail {
		os.Exit(1)
	}
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
