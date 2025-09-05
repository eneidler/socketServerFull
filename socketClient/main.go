package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Printf("Failed to connect to server: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to server!")
	fmt.Println("Commands: /quit (exit), /list (show users), /time (server time)")
	fmt.Println("Just type messages to chat with others")

	// Start goroutine to read messages from server
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	// Read input from user and send to server
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		message := strings.TrimSpace(scanner.Text())

		if message == "" {
			continue
		}

		// Send message to server
		_, err := conn.Write([]byte(message + "\n"))
		if err != nil {
			fmt.Printf("Error sending message: %v\n", err)
			break
		}

		// If user typed /quit, exit
		if message == "/quit" {
			break
		}
	}

	fmt.Println("Disconnected from server")
}
