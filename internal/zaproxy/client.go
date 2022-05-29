package zaproxy

import (
	"fmt"

	"github.com/1azunna/zap-api-go/zap"
	"github.com/sirupsen/logrus"
)

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
