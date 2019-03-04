package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"

	"github.com/amitbet/teleporter/agent"
	"github.com/amitbet/teleporter/logger"
)

func readConfig(file string) (*agent.AgentConfig, error) {
	clientConfigStr, err := ioutil.ReadFile(file)
	if err != nil {
		logger.Error("Client connect, failed while reading client header: %s\n", err)
		return nil, err
	}

	//logger.Debug("client connected, read client config string: ", clientConfigStr)
	cconfig := agent.AgentConfig{}
	err = json.Unmarshal([]byte(clientConfigStr), &cconfig)
	if err != nil {
		logger.Error("Client connect, error unmarshaling clientConfig: %s\n", err)
		return nil, err
	}
	return &cconfig, nil
}

func writeConfig(file string, config *agent.AgentConfig) error {
	jstr, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		logger.Error("writeConfig: problem in netConfig json marshaling: ", err)
		return err
	}
	err = ioutil.WriteFile(file, jstr, 0644)
	//err = common.WriteString(conn, string(jstr))
	if err != nil {
		logger.Error("writeConfig: Problem in sending network config: ", err)
		return err
	}
	return nil
}

// FileExists checks if a file exists in the given path
func FileExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func main() {
	confFile := "./config.json"
	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) > 0 {
		confFile = argsWithoutProg[0]
	}

	if !FileExists(confFile) {
		host, _ := os.Hostname()
		//os.Create(confFile)
		conf := agent.AgentConfig{
			AuthenticateSocks5: false,
			NetworkConfiguration: agent.ClientConfig{
				ClientId: host,
				Mapping:  make(map[string]string),
			},
			Connections: []agent.TetherConfig{
				agent.TetherConfig{
					TargetPort:     10201,
					TargetHost:     "<RemoteHost Address Or IP>",
					ConnectionType: "tls",
					ConnectionName: "<Some name or description like: network Node #2, should have Id = HomeComputer>",
					Credentials:    "",
				},
			},
			Servers: []agent.ListenerConfig{
				agent.ListenerConfig{
					Port:      10101,
					Type:      "Socks5",
					LocalOnly: true,
				},
				agent.ListenerConfig{
					Port:      10102,
					Type:      "relayTcp",
					LocalOnly: false,
				},
			},
		}

		conf.NetworkConfiguration.Mapping["*"] = "local"

		err := writeConfig(confFile, &conf)
		if err != nil {
			logger.Error("error in writing config: ", err)
		}
		fmt.Println("A Configuration file 'config.json' was written, please edit it and relaunch!")
		return
	}

	cconf, err := readConfig(confFile)
	if err != nil {
		logger.Error("Problem while reading configuration: ", err)
		return
	}

	rtr := agent.NewRouter()
	rtr.NetworkConfig = &cconf.NetworkConfiguration
	rtr.AuthenticateSocks5 = cconf.AuthenticateSocks5

	//facilitate all connections
	for _, connConf := range cconf.Connections {
		targetURI := connConf.TargetHost + ":" + strconv.Itoa(connConf.TargetPort)

		// look for a proxy configuration in connection first, and afterwards in whole client conf (nil is ok if none exist)
		proxy := connConf.Proxy
		if proxy == nil {
			proxy = cconf.Proxy
		}
		err := rtr.Connect(targetURI, connConf.ConnectionType, proxy)
		if err != nil {
			logger.Error("Agent: failed to connect to "+targetURI+": ", err)
		}
		logger.Info("Agent connected with: "+connConf.ConnectionType+" to ", targetURI)
	}

	//run all server listerners:
	for _, listenConf := range cconf.Servers {
		err := rtr.Serve(strconv.Itoa(listenConf.Port), listenConf.Type, listenConf.LocalOnly)
		if err != nil {
			logger.Error("Agent: failed to run listener: ", err)
			return
		}
		logger.Info("Agent: "+listenConf.Type+" listentning on port:", listenConf.Port)
	}

	// wait for ctrl+c to exit
	var signalChannel chan os.Signal
	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	func() {
		<-signalChannel
	}()
}
