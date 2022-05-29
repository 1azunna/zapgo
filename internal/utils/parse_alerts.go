package utils

import (
	"github.com/1azunna/zapgo/internal/types"
	"github.com/mitchellh/mapstructure"
)

func getLevel(inputList []string, level string) int {

	for i, lev := range inputList {
		if lev == level {
			return i
		}
	}
	return 0
}

func GetAlerts(data map[string]interface{}, filters types.AlertFilters) [][]string {

	var alerts [][]string
	var input interface{}
	riskLevels := []string{0: "Informational", 1: "Low", 2: "Medium", 3: "High"}
	confLevels := []string{0: "Low", 1: "Medium", 2: "High", 3: "Confirmed"}

	cfg := &mapstructure.DecoderConfig{
		TagName: "json",
	}

	if err := mapstructure.Decode(data, &input); err != nil {
		panic(err)
	}

	inp := input.(map[string]interface{})
	alertsByRisk := inp["alertsByRisk"]
	int1 := alertsByRisk.([]interface{})

	for j := 0; j < len(int1); j++ {
		i := int1[j].(map[string]interface{})
		for _, w := range i {
			det := w.([]interface{})
			for l := 0; l < len(det); l++ {
				x := det[l].(map[string]interface{})

				for _, val := range x {
					details := val.([]interface{})

					for m := 0; m < len(details); m++ {
						n := details[m].(map[string]interface{})
						var output types.AlertDetails
						cfg.Result = &output
						decoder, _ := mapstructure.NewDecoder(cfg)
						if err := decoder.Decode(n); err != nil {
							panic(err)
						}

						addAlert := false
						risk := getLevel(riskLevels, output.Risk)
						rLevel := getLevel(riskLevels, filters.Risk)
						if risk >= rLevel {
							addAlert = true
						}
						conf := getLevel(confLevels, output.Confidence)
						confLevel := getLevel(confLevels, filters.Confidence)
						if conf >= confLevel && addAlert {
							alerts = append(alerts, []string{output.Name, output.Risk, output.Confidence, output.Url, output.Param})
						}
					}
				}
			}
		}
	}
	return alerts
}

func GetAlertsCount(data map[string]interface{}) types.AlertsCount {
	var count types.AlertsCount

	cfg := &mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &count,
	}
	decoder, _ := mapstructure.NewDecoder(cfg)
	if err := decoder.Decode(data); err != nil {
		panic(err)
	}

	return count
}
