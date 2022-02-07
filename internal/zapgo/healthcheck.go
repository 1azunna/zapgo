package zapgo

import (
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var retrySchedule = []time.Duration{
	5 * time.Second,
	10 * time.Second,
	15 * time.Second,
	30 * time.Second,
}

func (z *Zapgo) HealthCheck(url string) {
	var response *http.Response
	// Wait 10seconds before checking for liveness
	time.Sleep(10 * time.Second)
	for _, backoff := range retrySchedule {
		resp, err := http.Get(url)
		if err != nil {
			logrus.Warn("Container is not ready")
			logrus.Warnf("Retrying in %v", backoff)
			time.Sleep(backoff)
		} else if err == nil {
			response = resp
			logrus.Info("Container initialized successfully!")
			break
		} else {
			logrus.Fatal(err)
		}
		defer resp.Body.Close()

	}
	if response == nil {
		logrus.Error("Could not verify that zap is running.")
		os.Exit(1)
	}
}
