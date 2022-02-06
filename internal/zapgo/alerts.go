package zapgo

import (
	"github.com/mitchellh/mapstructure"
)

type Details struct {
	Confidence string `json:"confidence"`
	Id         string `json:"id"`
	Name       string `json:"name"`
	Param      string `json:"param"`
	Risk       string `json:"risk"`
	Url        string `json:"url"`
}

type AlertsCount struct {
	High          int `json:"high"`
	Medium        int `json:"medium"`
	Low           int `json:"low"`
	Informational int `json:"informational"`
}

type Filters struct {
	Confidence string
	Risk       string
}

func getLevel(inputList []string, level string) int {

	for i, lev := range inputList {
		if lev == level {
			return i
		}
	}
	return 0
}

func GetAlerts(data map[string]interface{}, filters Filters) [][]string {
	// Trying to use mappstructure to convert the input to struct
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
	// fmt.Println(data)

	inp := input.(map[string]interface{})
	alertsByRisk := inp["alertsByRisk"]
	int1 := alertsByRisk.([]interface{})
	// fmt.Println(int1)
	for j := 0; j < len(int1); j++ {
		i := int1[j].(map[string]interface{})
		for _, w := range i {
			det := w.([]interface{})
			for l := 0; l < len(det); l++ {
				x := det[l].(map[string]interface{})
				// fmt.Println(x)
				for _, val := range x {
					details := val.([]interface{})
					// fmt.Println(details)
					for m := 0; m < len(details); m++ {
						n := details[m].(map[string]interface{})
						var output Details
						cfg.Result = &output
						decoder, _ := mapstructure.NewDecoder(cfg)
						if err := decoder.Decode(n); err != nil {
							panic(err)
						}
						// fmt.Printf("%+v\n", output)
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

func GetAlertsCount(data map[string]interface{}) AlertsCount {
	var count AlertsCount

	cfg := &mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &count,
	}
	decoder, _ := mapstructure.NewDecoder(cfg)
	if err := decoder.Decode(data); err != nil {
		panic(err)
	}
	// fmt.Printf("%+v\n", count)
	return count
}
