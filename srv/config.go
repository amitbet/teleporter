package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/amitbet/teleporter/logger"
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
		logger.Error("Failed to load config", err)
		return
	}

	conf = &config{}
	err = json.Unmarshal(jsonBlob, conf)
	if err != nil {
		logger.Error("Failed to parse config", err)
		return
	}

	return
}
