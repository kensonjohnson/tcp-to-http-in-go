package server

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"tcp-to-http/internal/request"
	"tcp-to-http/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		handler:  handler,
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
			log.Printf("Error in listen: %s\n", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		herr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		herr.Write(conn)
		return
	}

	buf := bytes.NewBuffer([]byte{})
	herr := s.handler(buf, req)
	if herr != nil {
		herr.Write(conn)
		return
	}

	b := buf.Bytes()
	response.WriteStatusLine(conn, response.StatusOK)
	h := response.GetDefaultHeaders(len(b))
	response.WriteHeaders(conn, h)
	conn.Write(b)
}
