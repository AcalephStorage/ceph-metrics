package main

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
)

const (
	OK          = 200
	WARN        = 429
	UNAVAILABLE = 503
)

// http api should possible run on all instances but redirects it to the leader

func startHttpServer(bindAddress string) {
	http.HandleFunc("/ceph/", cephHealth)

	logrus.Infoln("Http Check Server started.")
	http.ListenAndServe(bindAddress, nil)
}

func cephHealth(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	health := healthMap[req.RequestURI]
	logrus.Println(req.RequestURI)
	var response int

	switch health.status {
	case HealthOk:
		response = OK
	case HealthWarn:
		response = WARN
	default:
		response = UNAVAILABLE
	}
	res.WriteHeader(response)
	res.Write([]byte(health.message + "\n"))
	fmt.Println(health.message)
}
