package zaproxy

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/1azunna/zap-api-go/zap"
	"github.com/1azunna/zapgo/internal/defaults"
	"github.com/1azunna/zapgo/internal/utils"
	"github.com/sirupsen/logrus"
)

func getPlan(file string, path string) string {
	// Check for automation file
	if _, err := os.Stat(filepath.Join(path, file)); errors.Is(err, os.ErrNotExist) {
		logrus.Fatalf("The specified file %s does not exist.", filepath.Join(path, file))
	}
	filename := filepath.Join("/zap/wrk/", file)
	return filename
}

func ImportPolicy(file string, path string, zapClient zap.Interface) {
	//Check for policy file
	if _, err := os.Stat(filepath.Join(path, file)); len(file) > 0 && errors.Is(err, os.ErrNotExist) {
		logrus.Fatalf("The specified file %s does not exist.", filepath.Join(path, file))
	} else if len(file) > 0 && err == nil {
		policyfile := filepath.Join("/zap/wrk/", file)
		resp, err := zapClient.Ascan().ImportScanPolicy(policyfile)
		if err != nil {
			panic(err)
		}
		// fmt.Println(resp)
		logrus.Debugf("importing scan policy %s : %s", filepath.Join(path, file), resp["Result"])
	}
}

func RunPlan(file string, path string, zapClient zap.Interface) string {
	planfile := getPlan(file, path)
	logrus.Infof("Running automation framework with file %s", planfile)

	yamlData, err := ioutil.ReadFile(file)
	if err != nil {
		panic(fmt.Sprintf("Could not parse yaml file %s \n%s", file, err))
	}
	// Will use this later to print results.
	utils.PrintContexts(yamlData)

	plan, err := zapClient.Automation().RunPlan(planfile)
	if err != nil {
		panic(err)
	}
	for k, v := range plan {
		logrus.Debug(fmt.Sprintf("%s:%s", k, v))
		if k == "code" {
			logrus.Error(plan["message"].(string))
		}
	}

	planId := plan["planId"].(string)

	return planId
}

func GetPlanStatus(planId string, zapClient zap.Interface) {
	c := time.Tick(10 * time.Second)
	index := 0
	maxIndexSecs := defaults.MaxScanDurationInMins * 60
	maxIndex := maxIndexSecs / 10

	for range c {
		finished := false
		//Get the plan progress
		resp, err := zapClient.Automation().PlanProgress(planId)
		if err != nil {
			panic(err)
		}
		// Check if the status is finished
		index = index + 1
		if index == maxIndex {
			logrus.Error("Plan Timeout Exceeded")
			zapClient.Core().Shutdown()
			os.Exit(1)
		}
		// TODO: Implement check for plan errors
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

}
