package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

const (
	ExitString string = "-close"
	TCP        string = "tcp"
)

var mu sync.Mutex

type Server struct {
	listener net.Listener
	address  string
	clients  map[net.Conn]bool
}

type Client struct {
	conn     net.Conn
	username string
}

func (s *Server) BroadcastMessage(msg string, sender net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	for client := range s.clients {
		if client != sender {
			_, err := client.Write([]byte(msg))
			if err != nil {
				log.Printf("could not send message to client: %s", sender.RemoteAddr())
				sender.Close()
				delete(s.clients, sender)
			}
		}
	}
}

func (s *Server) RemoveClient(c net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	delete(s.clients, c)
}

var TCPServer Server

func main() {
	var port int
	flag.IntVar(&port, "port", 8000, "specifies port value")

	flag.Parse()

	address := ":" + strconv.Itoa(port)

	server := CreateTCPServer(address)
	defer server.listener.Close()
	log.Printf("Server is listening on port: %d", port)

	for {
		conn, err := server.listener.Accept()
		if err != nil {
			log.Printf("could not accept connection, %v", err)
			continue
		}
		mu.Lock()
		server.clients[conn] = true
		mu.Unlock()

		go handleConnection(conn, server)
	}

}

func handleConnection(conn net.Conn, server *Server) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	_, err := conn.Write([]byte("Connected to: " + conn.LocalAddr().String() + " To exit write: -close\n"))
	if err != nil {
		log.Printf("Could not write to client: %v", err)
		return
	}
	for scanner.Scan() {
		text := scanner.Text()
		log.Printf("msg: %s\n", text)
		msg := strings.TrimSpace(text)
		if msg == ExitString {
			log.Printf("Connection closed with client: %s", conn.RemoteAddr())
			server.BroadcastMessage(fmt.Sprintf("User: %s exited", conn.RemoteAddr().String()), conn)
			server.RemoveClient(conn)
			return
		}
		msg = conn.RemoteAddr().String() + ": " + msg + "\n"
		server.BroadcastMessage(msg, conn)
		if err != nil {
			log.Printf("Could not write to client: %v", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Connection error: %v", err)
	}
	log.Println("Closed connection: ", conn.RemoteAddr().String())
}

func CreateTCPServer(address string) *Server {
	server := Server{address: address, clients: make(map[net.Conn]bool)}
	listener, err := net.Listen(TCP, address)
	if err != nil {
		log.Fatalf("could not start server: %+v", err)
	}
	server.listener = listener

	return &server
}
