package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

func CurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
	}
	return dir
}
