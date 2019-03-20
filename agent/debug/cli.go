package main

import (
	"os"
	"os/signal"

	"github.com/amitbet/teleporter/agent"
)

func main() {
	rtr1 := agent.NewRouter()
	relayPass := agent.GenerateRandomString(32)
	conf1 := agent.ListenerConfig{
		Port:              10101,
		Type:              "relayTcp",
		LocalOnly:         false,
		UseAuthentication: true,
		AuthorizedClients: map[string]string{
			rtr1.NetworkConfig.ClientId + "2": relayPass,
		},
	}
	rtr1.Serve(conf1)
	rtr1.NetworkConfig.ClientId = rtr1.NetworkConfig.ClientId + "1"
	rtr1.NetworkConfig.Mapping["*"] = "local"

	rtr2 := agent.NewRouter()
	rtr2.AuthenticateSocks5 = false
	relayPass2 := agent.GenerateRandomString(32)
	conf2relay := agent.ListenerConfig{
		Port:              10201,
		Type:              "relayTcp",
		LocalOnly:         false,
		UseAuthentication: true,
		AuthorizedClients: map[string]string{
			"relayUser": relayPass2,
		},
	}
	conf2socks := agent.ListenerConfig{
		Port:              10202,
		Type:              "socks5",
		LocalOnly:         true,
		UseAuthentication: false,
	}

	rtr2.Serve(conf2relay)
	rtr2.Serve(conf2socks)
	rtr2.NetworkConfig.ClientId = rtr2.NetworkConfig.ClientId + "2"

	// create a mapping to send all google domains through the relay:
	rtr2.NetworkConfig.Mapping["*google*"] = rtr1.NetworkConfig.ClientId
	rtr2.NetworkConfig.Mapping["*"] = "local"

	// create a tether between router1 & router2
	rtr2.Connect(
		&agent.TetherConfig{
			TargetPort:     10101,
			TargetHost:     "localhost",
			ConnectionType: "tls",
			ConnectionName: "stam1",
			ClientPassword: relayPass,
		},
		10,
	)

	var signal_channel chan os.Signal
	signal_channel = make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	func() {
		<-signal_channel
	}()
}
