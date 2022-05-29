package zaproxy

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/1azunna/zap-api-go/zap"
	"github.com/1azunna/zapgo/internal/types"
	"github.com/1azunna/zapgo/internal/utils"
	"github.com/sirupsen/logrus"
)

func GetScanResults(file string, planId string, filters types.AlertFilters, display string, qualityGate string, zapClient zap.Interface) {

	yamlData, err := ioutil.ReadFile(file)
	if err != nil {
		panic(fmt.Sprintf("Could not parse yaml file %s \n%s", file, err))
	}

	//Get contexts
	var alertsList [][]string
	var totalCount types.RiskCount
	contexts := utils.GetContexts(yamlData)

	for _, v := range contexts {
		for _, url := range v.Urls {
			zapAlerts, err := zapClient.Alert().AlertsByRisk(url, "true")
			if err != nil {
				panic(err)
			}
			// fmt.Println(zapAlerts)
			alerts := utils.GetAlerts(zapAlerts, filters)
			for i := 0; i < len(alerts); i++ {
				alertsList = append(alertsList, alerts[i])
			}
			zapAlertsCount, err := zapClient.Alert().AlertCountsByRisk(url, "true")
			if err != nil {
				panic(err)
			}
			count := utils.GetAlertsCount(zapAlertsCount)
			totalCount.High = append(totalCount.High, count.High)
			totalCount.Medium = append(totalCount.Medium, count.Medium)
			totalCount.Low = append(totalCount.Low, count.Low)
			totalCount.Informational = append(totalCount.Informational, count.Informational)

			if display == "Sites" {
				fmt.Printf("\n Alerts discovered on Site: %s \n", url)
				utils.PrintAlerts(alertsList)
			}
		}
		if display == "Contexts" {
			fmt.Printf("\n Alerts discovered on Context: %s \n", v.Name)
			utils.PrintAlerts(alertsList)
		}
	}

	if display == "All" {
		utils.PrintAlerts(alertsList)
	}

	totalAlerts := utils.CountSum(totalCount.High) + utils.CountSum(totalCount.Medium) + utils.CountSum(totalCount.Low) + utils.CountSum(totalCount.Informational)
	logrus.Infof("A total of %s alerts were found. See zap.log for zap output.", strconv.Itoa(totalAlerts))
	// Set exit code 1 if fail is specified
	utils.SetExitCode(totalCount, qualityGate)
}
