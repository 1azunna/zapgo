package zaproxy

import (
	"errors"
	"fmt"
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
			logrus.Fatalln(err)
		}
		// fmt.Println(resp)
		logrus.Debugf("importing scan policy %s : %s", filepath.Join(path, file), resp["Result"])
	}
}

func RunPlan(file string, path string, zapClient zap.Interface) string {
	planfile := getPlan(file, path)
	logrus.Infof("Running automation framework with file %s", planfile)

	yamlData, err := os.ReadFile(file)
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
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	m := (defaults.MaxScanDurationInMins * 60) / 10
	done := make(chan bool)

	for i := -1; i < m; i++ {
		start := time.Now()
		duration := time.Since(start)

		go func() {
			//Get the plan progress
			resp, err := zapClient.Automation().PlanProgress(planId)
			if err != nil {
				logrus.Fatalln(err)
			}
			// TODO: Implement check for plan errors
			for k := range resp {
				if k == "finished" {
					done <- true
					return
				}
			}
		}()

		select {
		case <-done:
			logrus.Infoln("Automation plan complete. Scan took", duration/time.Second)
			return
		case <-ticker.C:
		}

		// Check if the status is finished
		if i >= m {
			logrus.Error("Plan Timeout Exceeded")
			_, err := zapClient.Core().Shutdown()
			if err != nil {
				logrus.Fatalln(err)
			}
			os.Exit(1)
		}
	}

}
