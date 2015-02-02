package network

import "net"

const (
	ipv4Mdns = "224.0.0.251"
	mdnsPort = 5353
)

var (
	ipv4Addr = &net.UDPAddr{IP: net.ParseIP(ipv4Mdns), Port: mdnsPort}
)

