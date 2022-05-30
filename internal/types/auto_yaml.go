package types

type ConfigBase struct {
	Env struct {
		Contexts []struct {
			Name string   `yaml:"name"`
			Urls []string `yaml:"urls"`
		}
	}
	Jobs []struct{} `yaml:"jobs"`
}

type Contexts []struct {
	Name string   `yaml:"name"`
	Urls []string `yaml:"urls"`
}
