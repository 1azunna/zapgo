package types

type AlertsCount struct {
	High          int `json:"high"`
	Medium        int `json:"medium"`
	Low           int `json:"low"`
	Informational int `json:"informational"`
}

type AlertDetails struct {
	Confidence string `json:"confidence"`
	Id         string `json:"id"`
	Name       string `json:"name"`
	Param      string `json:"param"`
	Risk       string `json:"risk"`
	Url        string `json:"url"`
}

type AlertFilters struct {
	Confidence string
	Risk       string
}
