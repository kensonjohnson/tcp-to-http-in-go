package server

import (
	"io"
	"tcp-to-http/internal/request"
	"tcp-to-http/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func (he HandlerError) Write(w io.Writer) error {
	response.WriteStatusLine(w, he.StatusCode)
	h := response.GetDefaultHeaders(len(he.Message))
	response.WriteHeaders(w, h)
	w.Write([]byte(he.Message))
	return nil
}
