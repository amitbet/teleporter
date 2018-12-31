package main

import (
	"log"
	"net/http"
)

func startHTTP(port string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("no comment")) // clientjson
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))

}
