package network

import (
	"net"
	"fmt"
	"sync"
	"golang.org/x/net/ipv4"
)

const (
	ipv4Mdns = "224.0.0.251"
	mdnsPort = 5353
)

var (
	ipv4Addr = &net.UDPAddr{IP: net.ParseIP(ipv4Mdns), Port: mdnsPort}
)

type Peer struct {
	ipv4UnicastConn *net.UDPConn
	ipv4Listener *net.UDPConn
	sendLock *sync.Mutex
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
		sendLock: &sync.Mutex{},
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
	pconn1 := ipv4.NewPacketConn(p.ipv4Listener)
	err1 := pconn1.SetMulticastInterface(iface)
	if err1 != nil {
		return err1
	}
	return nil
}

// multicast a query out
func (c *Peer) SendMulticast(m Message) {
	c.sendLock.Lock()
	defer c.sendLock.Unlock()
	c.ipv4UnicastConn.WriteToUDP(m, ipv4Addr)
}

// unicast a query out
func (c *Peer) SendUnicast(m Message, to net.Addr) {
	addr := to.(*net.UDPAddr)
	c.sendLock.Lock()
	defer c.sendLock.Unlock()
	c.ipv4UnicastConn.WriteToUDP(m, addr)
}

func (c *Peer) ListenUnicast() {
	buf := make([]byte, 65536)
	for {
		n, from, err := c.ipv4UnicastConn.ReadFrom(buf)
		if err != nil {
			fmt.Printf("[ERR] client: Failed to read packet: %v", err)
			continue
		}
		c.RecvCh <- Response{buf[:n], from}
	}
}

func (s *Peer) ListenMulticast() {
	buf := make([]byte, 65536)
	for {
		n, from, err := s.ipv4Listener.ReadFrom(buf)
		if err != nil {
			continue
		}
		s.RecvCh <- Response{buf[:n], from}
	}
}
