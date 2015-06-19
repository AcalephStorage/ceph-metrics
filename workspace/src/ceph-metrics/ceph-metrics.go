package main

import (
	"gopkg.in/alecthomas/kingpin.v1"
)

type (
	HealthStatus int
	HealthUpdate struct {
		uri     string
		status  HealthStatus
		message string
	}
)

const (
	HealthUnknown  HealthStatus = iota
	HealthOk       HealthStatus = iota
	HealthWarn     HealthStatus = iota
	HealthCritical HealthStatus = iota

	CephHealthUri    = "/ceph/health"
	CephMonQuorumUri = "/ceph/mon/quorum"
)

var (
	cluster             = kingpin.Flag("cluster", "ceph cluster to connect to").Default("ceph").String()
	cephBinary          = kingpin.Flag("ceph-bin", "The ceph binary").Default("/usr/bin/ceph").String()
	bindAddress         = kingpin.Flag("bind", "address and port to bind the http server").Default(":8080").String()
	interval            = kingpin.Flag("interval", "interval when sending metrics").Default("5").Int()
	healthMap           = make(map[string]HealthUpdate)
	healthUpdateChannel = make(chan HealthUpdate)
)

func main() {
	kingpin.Version("1.0.0")
	kingpin.Parse()

	go processMetrics()
	go updateHealth(healthUpdateChannel)
	startHttpServer(*bindAddress)
}

func updateHealth(healthUpdateChannel chan HealthUpdate) {
	for {
		update := <-healthUpdateChannel
		healthMap[update.uri] = update
	}
}
