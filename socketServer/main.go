package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// Client represents a connected client
type Client struct {
	conn     net.Conn
	nickname string
	server   *Server
}

// Server manages all connected clients
type Server struct {
	clients map[net.Conn]*Client
	mutex   sync.RWMutex
	address string
}

// NewServer creates a new server instance
func NewServer(address string) *Server {
	return &Server{
		clients: make(map[net.Conn]*Client),
		address: address,
	}
}

// Start begins listening for connections
func (s *Server) Start() error {
	// Listen on TCP port
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer listener.Close()

	fmt.Printf("Socket server started on %s\n", s.address)

	// Accept connections in a loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		// Handle each client connection in a separate goroutine
		go s.handleClient(conn)
	}
}

// handleClient manages communication with a single client
func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()

	// Create new client
	client := &Client{
		conn:   conn,
		server: s,
	}

	// Send welcome message
	client.send("Welcome to the Go Socket Server!")
	client.send("Please enter your nickname:")

	// Read nickname
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		client.nickname = strings.TrimSpace(scanner.Text())
		if client.nickname == "" {
			client.nickname = "Anonymous"
		}
	}

	// Add client to server's client list
	s.addClient(client)

	// Notify others of new client
	s.broadcast(fmt.Sprintf("%s joined the chat", client.nickname), client)

	fmt.Printf("Client %s connected from %s\n", client.nickname, conn.RemoteAddr())

	// Listen for messages from this client
	for scanner.Scan() {
		message := strings.TrimSpace(scanner.Text())

		if message == "" {
			continue
		}

		// Handle special commands
		if strings.HasPrefix(message, "/quit") {
			break
		} else if strings.HasPrefix(message, "/list") {
			client.send(s.getClientList())
		} else if strings.HasPrefix(message, "/time") {
			client.send("Server time: " + time.Now().Format("15:04:05"))
		} else if strings.HasPrefix(message, "/kick") {
			s.handleKickClient(message, client)
		} else {
			// Broadcast regular message to all clients
			fullMessage := fmt.Sprintf("[%s]: %s", client.nickname, message)
			s.broadcast(fullMessage, client)
		}
	}

	// Client disconnected
	s.removeClient(client)
	s.broadcast(fmt.Sprintf("%s left the chat", client.nickname), nil)
	fmt.Printf("Client %s disconnected\n", client.nickname)
}

// send sends a message to this specific client
func (c *Client) send(message string) {
	_, err := c.conn.Write([]byte(message + "\n"))
	if err != nil {
		fmt.Printf("Error sending message to %s: %v\n", c.nickname, err)
	}
}

// addClient adds a client to the server's client list (thread-safe)
func (s *Server) addClient(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.clients[client.conn] = client
}

// removeClient removes a client from the server's client list (thread-safe)
func (s *Server) removeClient(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.clients, client.conn)
}

// kickClient forcibly kicks client from the server and notifies them (thread-safe)
func (s *Server) kickClient(nickname string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for conn, client := range s.clients {
		if client.nickname == nickname {
			client.send("You have been kicked!")

			delete(s.clients, conn)

			go func() {
				err := conn.Close()
				if err != nil {
					return
				}
			}()

			return true
		}
	}
	return false
}

// handleKickClient handles the request to kick a client from the server
func (s *Server) handleKickClient(message string, sender *Client) {
	messageSplit := strings.SplitN(message, " ", 2)
	if len(messageSplit) != 2 {
		sender.send("Usage: /kick <nickname>")
		return
	}

	targetNickname := messageSplit[1]

	kicked := s.kickClient(targetNickname)
	if kicked {
		sender.send(fmt.Sprintf("You have kicked %s", targetNickname))
		s.broadcast(fmt.Sprintf("%s was kicked by %s", targetNickname, sender.nickname), nil)
	}
}

// broadcast sends a message to all connected clients except the sender
func (s *Server) broadcast(message string, sender *Client) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for conn, client := range s.clients {
		// Don't send message back to sender
		if sender != nil && conn == sender.conn {
			continue
		}
		client.send(message)
	}
}

// getClientList returns a formatted list of connected clients
func (s *Server) getClientList() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if len(s.clients) == 0 {
		return "No clients connected"
	}

	result := fmt.Sprintf("Connected clients (%d):", len(s.clients))
	for _, client := range s.clients {
		result += fmt.Sprintf("\n- %s", client.nickname)
	}
	return result
}

func main() {
	// Create and start the server
	server := NewServer(":8080")

	fmt.Println("Starting Go Socket Server...")
	if err := server.Start(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
