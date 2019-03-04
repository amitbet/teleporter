# Teleporter
The Teleporter project allows you to create your own network,
it is a simple, one binary super-server that breakes through network bounderies

*Teleporter, while fully functional is still a beta. (Use at your own discression)*

## Features:
* Onboards traffic through a Socks5 server (+ VPN support in the future) which can be configured on a browser or on the whole OS
* Creates strong bi-directional connections to other teleporter instances that can traverse network firewalls
* Routes traffic between such teleporter nodes by following simple wildecard rules in the JSON config file
* Works on any port
* Selectively exposes specific IPs or Domain names in the network to connected teleport nodes
* Support for multipls transport protocols (**currently only TLS**, future work: dtls/udp, websocket?)
* Can be chained to create a multi-hop network, or any other network formation you desire.
* A powerfull multiplexor engine, allows all traffic to be sent over a finite number of connections (Thanks to Alan Shreve's muxado project)
* No slowdown for traffic that enters & exist locally (local socks5 connections), 
* No software lags for relays, only mandatory network lags

## Security:
* Inter-node connections (tethers) are TLS encrypted 
* socks5 connections can be password protected (although not encrypted)
* socks5 connections can be restricted to accept only from localhost
* **Authentication features are still TBD**

## Potential Uses:
* Stay connected to home equipment without port mapping
* Seamlessly RDP/VNC into computers on multiple firewalled networks to provide remote support
* Expose on-premise webservers to potential customers or cloud testing farms
* Create secure agent connections from customer sites to cloud services
* Bridge network gaps without help from your IT department
* Stay connected to work without VPN
* Use as a custom VPN to spoof your origin, protect your privacy and gain access to location based services

## TODO:
* KeepAlive messages for TCP
* Http proxy "CONNECT" support for outgoing tls connections
* Authentication between nodes (clientId/Secret generation and storage)
* DTLS realy (secure udp) support
* Reconnect closed connections
* implement High Availability by connecting multiple times through a LB util enough connections report containing a link to the requested target host.