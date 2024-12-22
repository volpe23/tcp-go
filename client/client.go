package main

import (
	"bufio"
	"fmt"
	"strings"

	// "io"
	"log"
	"net"
	"os"
)

const TCP string = "tcp"

func main() {
	args := os.Args[1:]
	var host string
	var port string

	if len(args) == 2 {
		host = args[0]
		port = args[1]
	}

	addr := net.JoinHostPort(host, port)
	tcpAddr, err := net.ResolveTCPAddr(TCP, addr)
	if err != nil {
		log.Fatalf("could not resolve tcp address: %s", err)
	}

	conn, err := net.Dial(TCP, tcpAddr.String())

	if err != nil {
		log.Fatalf("could not connect to: %s\n", addr)
	}
	defer conn.Close()

	// _, err = conn.Write([]byte("connected to " + tcpAddr.String()))
	if err != nil {
		log.Fatalf("could not write to connection: %v", err)
	}

	go readMessages(conn)
	log.Println("listening to connection...")
	scanner := bufio.NewScanner(os.Stdin)

	// _, err = conn.Write([]byte("joined the room"))
	if err != nil {
		log.Fatalln("could not send message")
	}

	for scanner.Scan() {
		msg := strings.TrimSpace(scanner.Text())

		if msg == "-close" {
			break
		}
		if len(msg) > 0 {
			_, err := conn.Write([]byte(msg + "\n"))
			// fmt.Println(msg, i)
			if err != nil {
				fmt.Println("could not send message")
				break
			}
		}
	}
	fmt.Println("Exited")

}

func readMessages(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		fmt.Println(msg)
	}
}
