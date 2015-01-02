package server

import (
	"net"
	"lanfile/network/message"
)

const (
	ipv4_mdns = "224.0.0.251"
	mdns_port = 5353
)
var (
	ipv4_addr = &net.UDPAddr{IP: net.ParseIP(ipv4_mdns), Port: mdns_port}
)

type Server struct {
	ipv4_listener *net.UDPConn
	Recv_ch chan message.Response
}

func NewServer(comm chan message.Response) (*Server, error) {
	ipv4_listener, err := net.ListenMulticastUDP("udp4", nil, ipv4_addr)
	if ipv4_listener == nil {
		return nil, err
	}

	s := &Server {
		ipv4_listener: ipv4_listener,
		Recv_ch: comm,
	}

	return s, nil
}

func (s *Server) Listen() {
	buf := make([]byte, 65536)
	for {
		n, from, err := s.ipv4_listener.ReadFrom(buf)
		if err != nil {
			continue
		}
		s.Recv_ch <- message.Response{message.FromJson(buf[:n]), from}
	}
}

func (s *Server) SendUnicast(from net.Addr, msg message.Message) error {
	addr := from.(*net.UDPAddr)
	_, err := s.ipv4_listener.WriteToUDP(message.ToJson(msg), addr)
	return err
}
