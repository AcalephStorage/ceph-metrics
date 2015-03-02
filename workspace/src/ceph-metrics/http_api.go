package main

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func startHttpServer(bindAddress string) {
	r := mux.NewRouter()
	r.HandleFunc("/osd/stat", osdStat)

	http.Handle("/", r)

	logrus.Infoln("Http Check Server started.")
	http.ListenAndServe(bindAddress, nil)
}

func osdStat(res http.ResponseWriter, req *http.Request) {

}
