package zapgo

import (
	"fmt"

	"github.com/1azunna/zap-api-go/zap"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

// NewClient returns an object to communicate with the Docker Engine API.
func NewClient() *client.Client {

	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		logrus.Errorf("Could not create a docker client due to %v", err)
	}
	return client
}

// ZAP Http Client for making API requests
func ZapClient(baseUrl string) zap.Interface {
	cfg := zap.Config{
		Base:      fmt.Sprintf("%s/JSON/", baseUrl),
		BaseOther: fmt.Sprintf("%s/OTHER/", baseUrl),
		Proxy:     baseUrl,
	}
	client, err := zap.NewClient(&cfg)
	if err != nil {
		logrus.Fatal(err)
	}
	return client
}
