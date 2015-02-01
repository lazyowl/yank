package messaging

import (
	"net"
	"fmt"
	"golang.org/x/net/ipv4"
)

type Client struct {
	ipv4UnicastConn  *net.UDPConn

	RecvCh chan Response	// send from client to app
}

func NewClient(comm chan Response) (*Client, error) {
	// create a unicast ipv4 listener (listening on all available interfaces 0.0.0.0)
	uconn4, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if uconn4 == nil {
		return nil, err
	}

	c := &Client {
		ipv4UnicastConn: uconn4,
		RecvCh: make(chan Response),
	}

	// TODO change to try any available interface which has an IP address assigned to it by a DHCP server
	vbox, err := net.InterfaceByName("vboxnet0")
	c.SetInterface(vbox)

	return c, nil
}

// used to set the hardware interface
func (c *Client) SetInterface(iface *net.Interface) error {
	// need this to allow packets to be sent to the multicast group
	p := ipv4.NewPacketConn(c.ipv4UnicastConn)
	err := p.SetMulticastInterface(iface)
	if err != nil {
		return err
	}
	return nil
}

// multicast a query out
func (c *Client) SendMulticast(m Message) {
	byteStream := Serialize(m)
	c.ipv4UnicastConn.WriteToUDP(byteStream, ipv4Addr)
}

// unicast a query out
func (c *Client) SendUnicast(m Message, addr *net.UDPAddr) {
	byteStream := Serialize(m)
	c.ipv4UnicastConn.WriteToUDP(byteStream, addr)
}

func (c *Client) ListenUnicast() {
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
