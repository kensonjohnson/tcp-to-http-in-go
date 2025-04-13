package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Println(line)
		}
		fmt.Println("Connection closed from", conn.RemoteAddr())
	}
}

func getLinesChannel(conn net.Conn) <-chan string {
	lines := make(chan string)

	go func() {
		defer conn.Close()
		defer close(lines)
		buffer := make([]byte, 8)
		currentLine := ""
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				if currentLine != "" {
					lines <- currentLine
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				return
			}
			// Remember that a buffer could have some leftover data, so we only
			// want to take the part that was written during this loop
			str := string(buffer[:n])
			parts := strings.Split(str, "\n")
			currentLine += parts[0]
			if len(parts) > 1 {
				lines <- currentLine
				currentLine = parts[1]
			}
		}
	}()

	return lines
}
