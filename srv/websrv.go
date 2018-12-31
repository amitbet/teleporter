package main

import (
	"net/http"

	"github.com/amitbet/teleporter/logger"
)

func startHTTP(port string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("no comment")) // clientjson
	})

	logger.Fatal("Error in http listener: ", http.ListenAndServe(":"+port, nil))

}
