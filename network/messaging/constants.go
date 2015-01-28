package messaging

import "net"

const (
	ipv4_mdns = "224.0.0.251"
	mdns_port = 5353
)

var (
	ipv4_addr = &net.UDPAddr{IP: net.ParseIP(ipv4_mdns), Port: mdns_port}
)

