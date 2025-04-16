// chat-server is a TCP-based chat server that supports both broadcast and private messages
package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// Global variables for managing the chat server
var (
	// clients maps a connection to its username
	clients = make(map[net.Conn]string)
	// nameToConn maps a username to its connection
	nameToConn = make(map[string]net.Conn)
	//nameToPass maps a username to its password
	nameToPass = make(map[string]string)
	// broadcast channel for sending messages to all clients
	broadcast = make(chan string)
	// mutex for synchronizing access to shared data
	mutex             = &sync.Mutex{}
	lastPrivateSender = make(map[string]string) // maps recipient username to last sender username
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
	reader := bufio.NewReader(conn)

	// First, handle registration/login
	conn.Write([]byte("Welcome! Please register or login:\n"))
	conn.Write([]byte("1. To register: /register <username> <password>\n"))
	conn.Write([]byte("2. To login: /login <username> <password>\n"))

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message:", err)
			return
		}
		message = strings.TrimSpace(message)

		if strings.HasPrefix(message, "/register") {
			handleRegisterCommand(conn, message)
			break
		} else if strings.HasPrefix(message, "/login") {
			handleLoginCommand(conn, message)
			break
		} else {
			conn.Write([]byte("Please register or login first.\n"))
		}
	}

	// Get client's name after successful registration/login
	conn.Write([]byte("Enter your display name: "))
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
	broadcast <- fmt.Sprintf("\033[33m%s has joined the chat\033[0m\n", name)

	// Handle client messages
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		message = strings.TrimSpace(message)

		// Handle any commands, continue if a command was processed
		if handleCommand(conn, message) {
			continue
		}

		// Broadcast the message to all clients
		broadcast <- fmt.Sprintf("\033[34m%s: %s\033[0m\n", name, message)
	}

	// Clean up when client disconnects
	mutex.Lock()
	delete(clients, conn)
	delete(nameToConn, name)
	mutex.Unlock()
	broadcast <- fmt.Sprintf("\033[33m%s has left the chat\033[0m\n", name)
	conn.Close()
}

func checkRegistration(conn net.Conn) bool {
	mutex.Lock()
	defer mutex.Unlock()
	_, exists := clients[conn]
	return exists
}

func handleRegisterCommand(conn net.Conn, message string) {
	parts := strings.SplitN(message, " ", 3) // Changed to 3 to properly split username and password
	if len(parts) != 3 || strings.TrimSpace(parts[1]) == "" || strings.TrimSpace(parts[2]) == "" {
		conn.Write([]byte("Usage: /register <username> <password>\n"))
		return
	}
	username := strings.TrimSpace(parts[1])
	password := strings.TrimSpace(parts[2])

	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		conn.Write([]byte("Error registering user. Please try again.\n"))
		return
	}

	mutex.Lock()
	clients[conn] = username
	nameToConn[username] = conn
	nameToPass[username] = string(hashedPassword)
	mutex.Unlock()
	conn.Write([]byte(fmt.Sprintf("Welcome, %s! You can now start chatting.\n", username)))
}

// handleCommand handles any commands from the client
func handleCommand(conn net.Conn, message string) bool {
	//register command
	if strings.HasPrefix(message, "/register") {
		handleRegisterCommand(conn, message)
		return true
	}
	// /users command
	if strings.HasPrefix(message, "/users") {
		handleUsersCommand(conn)
		return true
	}
	// /private command
	if strings.HasPrefix(message, "/private") {
		handlePrivateMessage(conn, message)
		return true
	}

	// /reply command
	if strings.HasPrefix(message, "/reply") {
		handleReplyCommand(conn, message)
		return true
	}
	return false
}

// handleReplyCommand allows replying to the last private sender
func handleReplyCommand(conn net.Conn, message string) {
	mutex.Lock()
	username := clients[conn]
	lastSender, ok := lastPrivateSender[username]
	mutex.Unlock()
	if !ok {
		conn.Write([]byte("No private messages to reply to.\n"))
		return
	}
	parts := strings.SplitN(message, " ", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		conn.Write([]byte("Usage: /reply <message>\n"))
		return
	}
	msg := parts[1]
	handlePrivateMessage(conn, fmt.Sprintf("/private %s %s", lastSender, msg))
}

// handleUsersCommand handles the /users command
func handleUsersCommand(conn net.Conn) {
	mutex.Lock()
	for _, name := range clients {
		conn.Write([]byte("\033[90m" + name + "\033[0m\n"))
	}
	mutex.Unlock()
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

// handleLoginCommand handles user login
func handleLoginCommand(conn net.Conn, message string) {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) != 3 {
		conn.Write([]byte("Usage: /login <username> <password>\n"))
		return
	}
	username := strings.TrimSpace(parts[1])
	password := strings.TrimSpace(parts[2])

	mutex.Lock()
	hashedPassword, exists := nameToPass[username]
	mutex.Unlock()

	if !exists {
		conn.Write([]byte("User not found. Please register first.\n"))
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		conn.Write([]byte("Invalid password.\n"))
		return
	}

	conn.Write([]byte(fmt.Sprintf("Welcome back, %s!\n", username)))
}
