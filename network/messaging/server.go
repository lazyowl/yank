package messaging

import (
	"net"
)

type Server struct {
	ipv4_listener *net.UDPConn
	Recv_ch chan Response
}

func NewServer(comm chan Response) (*Server, error) {
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
		s.Recv_ch <- Response{FromJson(buf[:n]), from}
	}
}

func (s *Server) SendUnicast(from net.Addr, msg Message) error {
	addr := from.(*net.UDPAddr)
	_, err := s.ipv4_listener.WriteToUDP(ToJson(msg), addr)
	return err
}
