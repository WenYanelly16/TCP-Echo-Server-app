package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var (
	port    int
	timeout time.Duration
)

func init() {
	flag.IntVar(&port, "port", 4000, "Port to listen on")
	flag.DurationVar(&timeout, "timeout", 30*time.Second, "Client inactivity timeout")
	flag.Parse()
}

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
	logFile, err := os.Create(fmt.Sprintf("%s.log", strings.ReplaceAll(clientAddr, ":", "_")))
	if err != nil {
		log.Printf("Error creating log file for %s: %v", clientAddr, err)
		return
	}
	defer logFile.Close()
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
	activity := make(chan bool)
	defer close(activity)

	go func() {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		for {
			select {
			case <-activity:
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(timeout)
			case <-timer.C:
				writer.WriteString("Connection timed out due to inactivity\n")
				writer.Flush()
				conn.Close()
				return
			}
		}
	}()
}
