package messaging

import (
	"net"
	"fmt"
	"golang.org/x/net/ipv4"
)

type Client struct {
	ipv4_unicast_conn  *net.UDPConn

	Recv_ch chan Response	// send from client to app
}

func NewClient(comm chan Response) (*Client, error) {
	// create a unicast ipv4 listener (listening on all available interfaces 0.0.0.0)
	uconn4, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if uconn4 == nil {
		return nil, err
	}

	c := &Client {
		ipv4_unicast_conn: uconn4,
		Recv_ch: make(chan Response),
	}

	vbox, err := net.InterfaceByName("vboxnet0")
	c.SetInterface(vbox)

	return c, nil
}

// used to set the hardware interface
func (c *Client) SetInterface(iface *net.Interface) error {
	// need this to allow packets to be sent to the multicast group
	p := ipv4.NewPacketConn(c.ipv4_unicast_conn)
	err := p.SetMulticastInterface(iface)
	if err != nil {
		return err
	}
	return nil
}

// multicast a query out
func (c *Client) SendMulticast(m Message) {
	byteStream := ToJson(m)
	c.ipv4_unicast_conn.WriteToUDP(byteStream, ipv4_addr)
}

func (c *Client) ListenUnicast() {
	buf := make([]byte, 65536)
	for {
		n, from, err := c.ipv4_unicast_conn.ReadFrom(buf)
		if err != nil {
			fmt.Printf("[ERR] mdns: Failed to read packet: %v", err)
			continue
		}
		c.Recv_ch <- Response{FromJson(buf[:n]), from}
	}
}
