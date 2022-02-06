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
	// Wait 10seconds before checking for liveness
	time.Sleep(10 * time.Second)
	for _, backoff := range retrySchedule {
		resp, err := http.Get(url)
		if err == nil {
			logrus.Info("Container initialized successfully!")
			break
		}
		defer resp.Body.Close()

		logrus.Warn("Container is not ready")
		logrus.Warnf("Retrying in %v", backoff)
		time.Sleep(backoff)
	}
	resp, _ := http.Get(url)
	if resp == nil {
		logrus.Error("Could not reach host")
		os.Exit(1)
	}
	defer resp.Body.Close()

}
