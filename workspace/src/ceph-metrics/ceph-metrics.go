package main

import (
	"gopkg.in/alecthomas/kingpin.v1"
)

func main() {
	bindAddress := kingpin.Flag("bind", "address and port to bind the http server").Default(":8080").String()
	sendMetrics := kingpin.Flag("send-metrics", "enable sending of metrics to graphite/influxdb").Default("false").Bool()
	graphitePort := kingpin.Flag("graphite-port", "graphite port to send metrics to").Default("2003").Int()
	interval := kingpin.Flag("interval", "interval when sending metrics").Default("5").Int()
	kingpin.Version("1.0.0")
	kingpin.Parse()
	if *sendMetrics {
		go processMetrics(*interval, *graphitePort)
	}
	startHttpServer(*bindAddress)
}
