package main

import (
	"os"
	"os/signal"

	"github.com/amitbet/teleporter/agent"
)

func main() {
	// rtr1 := agent.NewRouter()
	// rtr1.Serve("10101", "relayTcp")

	// rtr2 := agent.NewRouter()
	// rtr2.Serve("10202", "relayTcp")
	// rtr2.Serve("10203", "socks5")
	// rtr2.Connect("localhost:10101", "tls")

	rtr1 := agent.NewRouter()
	rtr1.Serve("10201", "relayTcp", false)
	rtr1.NetworkConfig.ClientId = rtr1.NetworkConfig.ClientId + "1"

	// create a mapping to allow all domains to execute locally through this router
	rtr1.NetworkConfig.Mapping["*"] = "local"

	rtr2 := agent.NewRouter()
	rtr2.AuthenticateSocks5 = false
	//rtr2.Serve("10102", "relayTcp", false)
	rtr2.Serve("10101", "socks5", false)
	rtr2.NetworkConfig.ClientId = rtr2.NetworkConfig.ClientId + "2"

	// create a mapping to send all google domains through the relay:
	rtr2.NetworkConfig.Mapping["*google*"] = rtr1.NetworkConfig.ClientId
	rtr2.NetworkConfig.Mapping["*"] = "local"

	// create a tether between router1 & router2
	rtr2.Connect("localhost:10201", "tls")

	var signal_channel chan os.Signal
	signal_channel = make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	func() {
		<-signal_channel
	}()
}
