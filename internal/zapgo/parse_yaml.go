package zapgo

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ConfigFile struct {
	Env struct {
		Contexts []struct {
			Name string   `yaml:"name"`
			Urls []string `yaml:"urls"`
		}
	}
	Jobs []struct{} `yaml:"jobs"`
}

type Contexts []map[string][]string

var contexts Contexts

func GetContexts(config []byte, logger Logger) Contexts {
	c := ConfigFile{}
	err := yaml.Unmarshal(config, &c)
	if err != nil {
		panic(err)
	}
	m := make(map[string][]string)
	for _, v := range c.Env.Contexts {
		m[v.Name] = v.Urls
		contexts = append(contexts, m)
		logger.Info(fmt.Sprintf("Context: %s, Urls: %s", v.Name, v.Urls))
	}
	return contexts
}
