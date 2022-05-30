package types

type Zaproxy struct {
	Image     string
	Hostname  string
	Container string
	Network   string
	Port      string
	Options   []string
}

type Newman struct {
	Image       string
	Container   string
	Collection  string
	Environment string
}
