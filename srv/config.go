package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type config struct {
	CnCPort       string
	HTTPPort      string
	SocksUsername string
	SocksPort     string
	SocksPass     string
}

func loadConfig() (conf *config, err error) {
	jsonBlob, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Println("Failed to load config", err)
		return
	}

	conf = &config{}
	err = json.Unmarshal(jsonBlob, conf)
	if err != nil {
		log.Println("Failed to parse config", err)
		return
	}

	return
}
