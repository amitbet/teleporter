# Teleporter
The Teleporter project allows you to create your own network,
it is a simple, one binary super-server that breaks through network boundaries

*Teleporter, while fully functional is still a beta. (Use at your own discression)*

## Features:
* Onboards traffic through a Socks5 server (+ VPN support in the future) which can be configured on a browser or on the whole OS
* Creates strong bi-directional connections (tethers) to other teleporter instances that can traverse network firewalls
* Combats "head of line" problems by having multiple connections in each tether
* Routes traffic between teleporter nodes by following simple wildecard rules in the config file
* Selectively exposes specific IPs or Domain names in the network to connected teleport nodes
* Support for multipls transport protocols (**currently only TLS**, future work: dtls/udp)
* Can be chained to create a multi-hop network, or any other network formation you desire.
* A powerfull multiplexor engine, allows all traffic to be sent over a finite number of connections (Thanks to Alan Shreve's muxado project)
* No slowdown for traffic that enters & exist locally (local socks5 connections)
* Works on any port
* No software lags for relays, only mandatory network lags
* Http proxy support for outgoing tls connections (using "CONNECT" like any normal https conn)
* Support for Http proxy authentication

## Security:
* Inter-node connections (tethers) are TLS encrypted 
* socks5 connections can be password protected (although not encrypted)
* socks5 connections can be restricted to accept only from localhost
* **Authentication features are still TBD**

## Potential Uses:
* Stay connected to home equipment without port mapping
* Seamlessly RDP/VNC into multiple networks for remote support
* Create secure agent connections from customer sites to cloud services
* Bridge network gaps without help from your IT department
* Stay connected to work without using a VPN
* Use as a custom VPN to spoof your origin, protect your privacy and gain access to location based services
* Expose on-premise webservers to potential customers or cloud testing farms through socks5 +auth

## Usage:
**Running release samples:**
1. Download the latest build of Teleporter
1. Use “run.bat” in any of the examples to run several nodes on localhost
1. Configure your browser to use the socks5 proxy now running on localhost:10101 to see it in action.

**Setting up a real network:**
1. Spin up some free tier cloud machine (GCP will give you $300 for a year, with no obligations, just by feeding in credit card info, AWS free tier machines are also an option)
1. Place a node on that machine & open the relay port to the public
1. Deploy teleport nodes on your favorite machines & configure (see below) to construct your own custom slice of internet!

**Running samples from code:**
1. go get github.com/amitbet/teleporter
1. cd agent/cli
1. go build .
1. cd ../../examples
1. Use "run.bat" in any of the example directories
1. Configure your browser to use the socks5 proxy now running on localhost:10101
1. Deploy on your favorite machines & configure to construct your own custom slice of internet!
 
## TODO:
* KeepAlive messages for TLS connections
* Add VPN support by using [gotun2socks](https://github.com/txthinking/gotun2socks) in a way similar to [brook](https://github.com/txthinking/brook)
* DTLS realy (secure udp) support
* Reconnect closed connections
* implement High Availability by connecting multiple times through a LB util enough connections report containing a link to the requested target host.
* Some embedded webUI (maybe experiment with [packr](https://github.com/gobuffalo/packr))
