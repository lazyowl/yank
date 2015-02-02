package network

import (
	"net"
	"fmt"
	"golang.org/x/net/ipv4"
)

type Peer struct {
	ipv4UnicastConn *net.UDPConn
	ipv4Listener *net.UDPConn
	RecvCh chan Response
}

func NewPeer() (*Peer, error) {
	uconn4, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if uconn4 == nil {
		return nil, err
	}
	ipv4Listener, err := net.ListenMulticastUDP("udp4", nil, ipv4Addr)
	if ipv4Listener == nil {
		return nil, err
	}
	p := &Peer {
		ipv4UnicastConn: uconn4,
		ipv4Listener: ipv4Listener,
		RecvCh: make(chan Response),
	}

	vbox, err := net.InterfaceByName("vboxnet0")
	p.SetInterface(vbox)
	return p, nil
}

// used to set the hardware interface
func (p *Peer) SetInterface(iface *net.Interface) error {
	// need this to allow packets to be sent to the multicast group
	pconn := ipv4.NewPacketConn(p.ipv4UnicastConn)
	err := pconn.SetMulticastInterface(iface)
	if err != nil {
		return err
	}
	return nil
}

// multicast a query out
func (c *Peer) SendMulticast(m Message) {
	byteStream := Serialize(m)
	c.ipv4UnicastConn.WriteToUDP(byteStream, ipv4Addr)
}

// unicast a query out
func (c *Peer) SendUnicast(m Message, to net.Addr) {
	addr := to.(*net.UDPAddr)
	byteStream := Serialize(m)
	c.ipv4UnicastConn.WriteToUDP(byteStream, addr)
}

func (c *Peer) ListenUnicast() {
	buf := make([]byte, 65536)
	for {
		n, from, err := c.ipv4UnicastConn.ReadFrom(buf)
		if err != nil {
			fmt.Printf("[ERR] client: Failed to read packet: %v", err)
			continue
		}
		c.RecvCh <- Response{Deserialize(buf[:n]), from}
	}
}

func (s *Peer) ListenMulticast() {
	buf := make([]byte, 65536)
	for {
		n, from, err := s.ipv4Listener.ReadFrom(buf)
		if err != nil {
			continue
		}
		s.RecvCh <- Response{Deserialize(buf[:n]), from}
	}
}
