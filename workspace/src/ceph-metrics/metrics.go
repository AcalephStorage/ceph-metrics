package main

import (
	"time"

	"github.com/Sirupsen/logrus"
)

func processMetrics(interval, graphitePort int) {
	logrus.Infoln("Metrics gathering started.")
	for {
		time.Sleep(time.Duration(interval) * time.Second)
		logrus.Infoln("Sending Metrics...")
	}
}
