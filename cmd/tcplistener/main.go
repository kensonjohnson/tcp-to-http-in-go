package main

import (
	"fmt"
	"log"
	"net"
	"tcp-to-http/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("Error opening TCP listener", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("There was an error opening TCP connection: ", err)
		}
		defer conn.Close()
		fmt.Println("Accepted connection from", conn.RemoteAddr())
		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Request line:")
		fmt.Println("- Method:", req.RequestLine.Method)
		fmt.Println("- Target:", req.RequestLine.RequestTarget)
		fmt.Println("- Version:", req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}
		fmt.Println("Connection closed from", conn.RemoteAddr())
	}
}
