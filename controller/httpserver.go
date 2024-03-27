package controller

import (
	"io/ioutil"
	"log"
	"net/http"
)

func ListenHTTP() {
	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(":5000", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		} else {
			WriteBodyResponse(body)
		}
	}
}
