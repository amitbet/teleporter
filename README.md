# Teleporter
Socks5 on steroids.

The Teleporter project allows you to create your own network
A simple, one binary super-server that breakes through network bounderies.

## Features:
* Onboards traffic through a Socks5 (+ VPN support in the future)
* Creates strong bi-directional connections to other tleporter instances that can traverse network firewalls
* Routes traffic between such teleporter nodes by following simple wildecard rules in it's JSON configuration
* A powerfull multiplexor engine, allows all traffic to be sent over a finite number of connections (Thanks to Alan Shreve's muxado project)
* Selectively exposes IPs/Domain names in the network to connected teleport nodes
* Support for multipls transport protocols (**currently only TLS**, future work: dtls, websockets)
* Works on any port

## Potential Uses:
* Stay connected to home equipment without port mapping
* Seamlessly RDP/VNC into computers on multiple networks
* Expose on-premise webservers to potential customers or cloud testing solutions
* Bridge network gaps without help from your IT department
* Stay connected to work without VPN