package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"tcp-to-http/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
	}

	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error in listener: %s\n", err)
			continue
		}
		fmt.Println("Connection accepted from:", conn.RemoteAddr())
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	response.WriteStatusLine(conn, 200)
	h := response.GetDefaultHeaders(0)
	response.WriteHeaders(conn, h)
	fmt.Println("Connection closed with:", conn.RemoteAddr())
}
