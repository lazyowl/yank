package messaging

import (
	"net"
	"fmt"
	"golang.org/x/net/ipv4"
)

type Client struct {
	ipv4_unicast_conn  *net.UDPConn
	ipv4_multicast_conn  *net.UDPConn

	Recv_ch chan Response	// send from client to app
}

func NewClient(iface string, comm chan Response) (*Client, error) {
	// create a unicast ipv4 listener (listening on all available interfaces 0.0.0.0)
	uconn4, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(iface), Port: 0})
	if uconn4 == nil {
		return nil, err
	}

	// create a multicast ipv4 listener (listening for UDP packets addressed to group address ipv4_addr)
	mconn4, err := net.ListenMulticastUDP("udp4", nil, ipv4_addr)
	if mconn4 == nil {
		return nil, err
	}

	c := &Client {
		ipv4_unicast_conn: uconn4,
		ipv4_multicast_conn: mconn4,
		Recv_ch: make(chan Response),
	}

	vbox, err := net.InterfaceByName(iface)
	c.SetInterface(vbox)

	return c, nil
}

// used to set the hardware interface
func (c *Client) SetInterface(iface *net.Interface) error {
	p := ipv4.NewPacketConn(c.ipv4_unicast_conn)
	p.SetMulticastLoopback(false)
	err := p.SetMulticastInterface(iface)
	if err != nil {
		return err
	}

	p = ipv4.NewPacketConn(c.ipv4_multicast_conn)
	err = p.SetMulticastInterface(iface)
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

func (c *Client) ListenMulticast() {
	buf := make([]byte, 65536)
	for {
		n, from, err := c.ipv4_multicast_conn.ReadFrom(buf)
		if err != nil {
			fmt.Printf("[ERR] mdns: Failed to read packet: %v", err)
			continue
		}
		c.Recv_ch <- Response{FromJson(buf[:n]), from}
	}
}
