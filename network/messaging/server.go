package messaging

import (
	"net"
)

type Server struct {
	ipv4Listener *net.UDPConn
	RecvCh chan Response
}

func NewServer(comm chan Response) (*Server, error) {
	ipv4Listener, err := net.ListenMulticastUDP("udp4", nil, ipv4Addr)
	if ipv4Listener == nil {
		return nil, err
	}

	s := &Server {
		ipv4Listener: ipv4Listener,
		RecvCh: comm,
	}

	return s, nil
}

func (s *Server) Listen() {
	buf := make([]byte, 65536)
	for {
		n, from, err := s.ipv4Listener.ReadFrom(buf)
		if err != nil {
			continue
		}
		s.RecvCh <- Response{Deserialize(buf[:n]), from}
	}
}

func (s *Server) SendUnicast(from net.Addr, msg Message) error {
	addr := from.(*net.UDPAddr)
	_, err := s.ipv4Listener.WriteToUDP(Serialize(msg), addr)
	return err
}
