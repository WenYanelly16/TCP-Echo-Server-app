package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":4000")
	log.Printf("Server listening on :%d", port)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Server listening on :4000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	clientAddr := conn.RemoteAddr().String()
	log.Printf("Client connected: %s", clientAddr)
	defer func() {
		conn.Close()
		log.Printf("Client disconnected: %s", clientAddr)
	}()
	defer conn.Close()
	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading from client:", err)
			return
		}
		_, err = conn.Write(buf[:n])
		if err != nil {
			fmt.Println("Error writing to client:", err)
		}
	}
}
