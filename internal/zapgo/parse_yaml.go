package zapgo

import (
	"github.com/sirupsen/logrus"
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

type Contexts []struct {
	Name string   `yaml:"name"`
	Urls []string `yaml:"urls"`
}

var contexts []map[string][]string

func GetContexts(config []byte) Contexts {
	c := ConfigFile{}
	err := yaml.Unmarshal(config, &c)
	if err != nil {
		panic(err)
	}
	return c.Env.Contexts
}

func PrintContexts(config []byte) []map[string][]string {
	c := ConfigFile{}
	err := yaml.Unmarshal(config, &c)
	if err != nil {
		panic(err)
	}
	m := make(map[string][]string)
	for _, v := range c.Env.Contexts {
		m[v.Name] = v.Urls
		contexts = append(contexts, m)
		logrus.Infof("Context: %s, Urls: %s", v.Name, v.Urls)
	}
	return contexts
}
