package main

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
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
	response := 404
	switch health.status {
	case HealthOk:
		response = 200
	case HealthWarn:
		response = 429
	default:
		response = 503
	}
	res.WriteHeader(response)
	res.Write([]byte(health.message + "\n"))
	fmt.Println(health.message)
}
