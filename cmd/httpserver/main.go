package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tcp-to-http/internal/request"
	"tcp-to-http/internal/response"
	"tcp-to-http/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, BasicHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func BasicHandler(w io.Writer, r *request.Request) *server.HandlerError {
	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    "Your problem is not my problem\n",
		}

	case "/myproblem":
		return &server.HandlerError{
			StatusCode: response.StatusInternalServerError,
			Message:    "Woopsie, my bad\n",
		}

	default:
		body := "All good, frfr\n"
		w.Write([]byte(body))
		return nil
	}
}
