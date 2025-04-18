package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// Global configuration variables
var (
	port    int       // Port number to listen on
	timeout time.Duration // Client inactivity timeout duration
)

// init function runs before main() to set up command line flags
func init() {
	// Define command line flags with default values and help text
	flag.IntVar(&port, "port", 4000, "Port to listen on")
	flag.DurationVar(&timeout, "timeout", 30*time.Second, "Client inactivity timeout")
	flag.Parse() // Parse the command line arguments
}

// main function - entry point of the server
func main() {
	// Create a TCP listener on the specified port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err) // Exit if listener fails
	}
	defer listener.Close() // Ensure listener is closed when main exits

	log.Printf("Server listening on :%d", port)

	// Main server loop - accepts incoming connections
	for {
		conn, err := listener.Accept() // Wait for and accept new connections
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue // Skip errors and keep accepting other connections
		}

		// Handle each connection in a separate goroutine for concurrency
		go handleConnection(conn)
	}
}

// handleConnection manages an individual client connection
func handleConnection(conn net.Conn) {
	clientAddr := conn.RemoteAddr().String() // Get client's address
	log.Printf("Client connected: %s", clientAddr)

	// Create a unique log file for this client (replace colons in address)
	logFile, err := os.Create(fmt.Sprintf("%s.log", strings.ReplaceAll(clientAddr, ":", "_")))
	if err != nil {
		log.Printf("Error creating log file for %s: %v", clientAddr, err)
		return
	}
	defer logFile.Close() // Ensure log file is closed when done

	// Deferred function to clean up when connection ends
	defer func() {
		conn.Close() // Close the connection
		log.Printf("Client disconnected: %s", clientAddr)
	}()

	// Create buffered reader/writer for the connection
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Channel to track client activity for timeout purposes
	activity := make(chan bool)
	defer close(activity) // Clean up channel when done

	// Goroutine to handle inactivity timeout
	go func() {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		for {
			select {
			case <-activity:
				// Reset timer on client activity
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(timeout)
			case <-timer.C:
				// Timeout reached - disconnect client
				writer.WriteString("Connection timed out due to inactivity\n")
				writer.Flush()
				conn.Close()
				return
			}
		}
	}()

	// Main message handling loop
	for {
		// Set read deadline to detect half-open connections
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		// Read message until newline
		message, err := reader.ReadString('\n')
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Handle read timeout (not necessarily inactivity)
				select {
				case activity <- true: // Signal activity if needed
					continue
				default:
					continue
				}
			}
			if err == io.EOF {
				return // Client closed connection normally
			}
			log.Printf("Error reading from %s: %v", clientAddr, err)
			return // Other errors - close connection
		}

		// Signal that client is active
		activity <- true

		// Log the raw message to console and file
		log.Printf("Message from %s: %s", clientAddr, strings.TrimSpace(message))
		fmt.Fprintf(logFile, "[%s] %s\n", time.Now().Format(time.RFC3339), strings.TrimSpace(message))

		// Clean the message by trimming whitespace
		message = strings.TrimSpace(message)

		// Handle empty message case
		if message == "" {
			writer.WriteString("Say something...\n")
			writer.Flush()
			continue
		}

		// Check for message length overflow
		if len(message) > 1024 {
			writer.WriteString("Error: Message too long (max 1024 bytes)\n")
			writer.Flush()
			continue
		}

		// Handle different message types
		switch {
		case strings.HasPrefix(message, "/"):
			// Process commands starting with /
			handleCommand(message, writer, conn)
		case strings.EqualFold(message, "hello"):
			// Special response for "hello"
			writer.WriteString("Hi there!\n")
		case strings.EqualFold(message, "bye"):
			// Special response for "bye" then disconnect
			writer.WriteString("Goodbye!\n")
			writer.Flush()
			return
		default:
			// Default behavior - echo the message
			writer.WriteString(message + "\n")
		}

		// Flush the writer buffer to ensure message is sent
		if err := writer.Flush(); err != nil {
			log.Printf("Error writing to %s: %v", clientAddr, err)
			return
		}
	}
}

// handleCommand processes special commands from the client
func handleCommand(cmd string, writer *bufio.Writer, conn net.Conn) {
	// Split command into parts (command and arguments)
	parts := strings.SplitN(cmd, " ", 2)
	command := strings.TrimSpace(parts[0])

	// Process different commands
	switch command {
	case "/time":
		// Return current server time
		writer.WriteString(time.Now().Format(time.RFC3339) + "\n")
	case "/quit":
		// Close the connection
		writer.WriteString("Closing connection\n")
		writer.Flush()
		conn.Close()
	case "/echo":
		// Echo back the provided message
		if len(parts) > 1 {
			writer.WriteString(parts[1] + "\n")
		} else {
			writer.WriteString("Usage: /echo <message>\n")
		}
	default:
		// Unknown command response
		writer.WriteString("Unknown command\n")
	}
}
