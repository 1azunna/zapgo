package utils

import (
	"github.com/1azunna/zapgo/internal/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var contexts []map[string][]string

func GetContexts(config []byte) types.Contexts {
	c := types.ConfigBase{}
	err := yaml.Unmarshal(config, &c)
	if err != nil {
		panic(err)
	}
	return c.Env.Contexts
}

func PrintContexts(config []byte) []map[string][]string {
	c := types.ConfigBase{}
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
