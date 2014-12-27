package server

import (
	"net"
	"fmt"
	"lanfile/network/message"
	//"golang.org/x/net/ipv4"
)

const (
	ipv4_mdns = "224.0.0.251"
	mdns_port = 5353
)
var (
	ipv4_addr = &net.UDPAddr{IP: net.ParseIP(ipv4_mdns), Port: mdns_port}
)

type Response struct {
	msg message.Message
	from net.Addr
}

type Server struct {
	ipv4_listener *net.UDPConn
	recv_ch chan Response
}

func NewServer(comm chan message.Message) (*Server, error) {
	ipv4_listener, err := net.ListenMulticastUDP("udp4", nil, ipv4_addr)
	if ipv4_listener == nil {
		return nil, err
	}

	s := &Server {
		ipv4_listener: ipv4_listener,
		recv_ch: make(chan Response),
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
		s.recv_ch <- Response{message.FromJson(buf[:n]), from}
	}
}

func (s *Server) StartLoop() {
	go s.Listen()
	for {
		select {
			case msg := <-s.recv_ch: {
				fmt.Println("server received ", msg.msg, msg.from)
				s.SendUnicast(msg.from, message.Message{"this is my response!"})
			}
		}
	}
}

func (s *Server) SendUnicast(from net.Addr, msg message.Message) error {
	addr := from.(*net.UDPAddr)
	_, err := s.ipv4_listener.WriteToUDP(message.ToJson(msg), addr)
	return err
}
