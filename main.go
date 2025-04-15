// chat-server is a TCP-based chat server that supports both broadcast and private messages
package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

// Global variables for managing the chat server
var (
	// clients maps a connection to its username
	clients = make(map[net.Conn]string)
	// nameToConn maps a username to its connection
	nameToConn = make(map[string]net.Conn)
	// broadcast channel for sending messages to all clients
	broadcast = make(chan string)
	// mutex for synchronizing access to shared data
	mutex = &sync.Mutex{}
)

// main starts the chat server
func main() {
	// Start listening on port 8080
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer ln.Close()

	// Start goroutines for handling messages
	go handleBroadcasting()     // Handle broadcast messages
	go processPrivateMessages() // Handle private messages

	fmt.Println("Server is running on port 8080")

	// Accept incoming connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}
		// Handle each client in a separate goroutine
		go handleClient(conn)
	}
}

// handleClient manages a single client connection
func handleClient(conn net.Conn) {
	// Get client's name
	conn.Write([]byte("Enter your name: "))
	reader := bufio.NewReader(conn)
	name, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading name:", err)
		return
	}
	name = strings.TrimSpace(name)

	// Add client to the server's client list
	mutex.Lock()
	clients[conn] = name
	nameToConn[name] = conn
	mutex.Unlock()

	// Notify everyone that a new client has joined
	broadcast <- fmt.Sprintf("%s has joined the chat\n", name)

	// Handle client messages
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		message = strings.TrimSpace(message)

		// Check if it's a private message
		if strings.HasPrefix(message, "/private") {
			handlePrivateMessage(conn, message)
			continue
		}

		// Broadcast the message to all clients
		broadcast <- fmt.Sprintf("%s: %s\n", name, message)
	}

	// Clean up when client disconnects
	mutex.Lock()
	delete(clients, conn)
	delete(nameToConn, name)
	mutex.Unlock()
	broadcast <- fmt.Sprintf("%s has left the chat\n", name)
	conn.Close()
}

// handleBroadcasting sends messages to all connected clients
func handleBroadcasting() {
	for message := range broadcast {
		mutex.Lock()
		for conn, _ := range clients {
			conn.Write([]byte(message))
		}
		mutex.Unlock()
	}
}
