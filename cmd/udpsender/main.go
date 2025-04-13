package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	UDPAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("Failed to get UDP Address")
	}

	conn, err := net.DialUDP("udp", nil, UDPAddr)
	if err != nil {
		log.Fatal("Failed to create UDP listener")
	}
	defer conn.Close()

	stdinReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, err := stdinReader.ReadString('\n')
		if err != nil {
			fmt.Println("There was an error in readString: ", err)
		}
		conn.Write([]byte(input))
	}
}
