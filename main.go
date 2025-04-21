// chat-server is a TCP-based chat server that supports both broadcast and private messages
package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// Global variables for managing the chat server
var (
	// clients maps a connection to its username
	clients = make(map[net.Conn]string)
	// nameToConn maps a username to its connection
	nameToConn = make(map[string]net.Conn)
	// displayNames tracks all used display names
	displayNames = make(map[string]bool)
	// broadcast channel for sending messages to all clients
	broadcast = make(chan string)
	// mutex for synchronizing access to shared data
	mutex             = &sync.Mutex{}
	lastPrivateSender = make(map[string]string) // maps recipient username to last sender username

	// Rate limiting for registration
	registerAttempts = make(map[string]int)       // IP -> attempt count
	registerTimes    = make(map[string]time.Time) // IP -> last attempt time
	registerMutex    = &sync.Mutex{}
)

// isRateLimited checks if an IP is rate limited for registration
func isRateLimited(ip string) bool {
	registerMutex.Lock()
	defer registerMutex.Unlock()

	now := time.Now()
	lastAttempt, exists := registerTimes[ip]

	// Reset counter if more than 1 minute has passed
	if exists && now.Sub(lastAttempt) > time.Minute {
		registerAttempts[ip] = 0
	}

	// Allow max 3 attempts per minute
	if registerAttempts[ip] >= 3 {
		return true
	}

	registerAttempts[ip]++
	registerTimes[ip] = now
	return false
}

// main starts the chat server
func main() {
	// Initialize database
	if err := initDB(); err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}
	defer closeDB()

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
	var username string
	var name string
	var authenticated bool

	// First, handle registration/login
	conn.Write([]byte("\033[1;36mWelcome to the Chat Server!\033[0m\n"))
	conn.Write([]byte("\033[1;32mPlease register or login:\033[0m\n"))
	conn.Write([]byte("\033[1;33m1. To register: /register <username> <password>\033[0m\n"))
	conn.Write([]byte("\033[1;33m2. To login: /login <username> <password>\033[0m\n"))

	for !authenticated {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message:", err)
			return
		}
		message = strings.TrimSpace(message)

		if strings.HasPrefix(message, "/register") {
			username = handleRegisterCommand(conn, message)
			if username != "" {
				authenticated = true
			}
		} else if strings.HasPrefix(message, "/login") {
			username = handleLoginCommand(conn, message)
			if username != "" {
				authenticated = true
			}
		} else if strings.HasPrefix(message, "/exit") {
			handleExitCommand(conn)
			return
		} else {
			conn.Write([]byte("\033[1;31mPlease register or login first.\033[0m\n"))
		}
	}

	// Get client's display name after successful registration/login
	for {
		conn.Write([]byte("\033[1;33mEnter your display name: \033[0m"))
		displayName, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading name:", err)
			return
		}
		displayName = strings.TrimSpace(displayName)

		// Check if display name is already taken
		mutex.Lock()
		if displayNames[displayName] {
			mutex.Unlock()
			conn.Write([]byte("\033[1;31mDisplay name already taken. Please choose another.\033[0m\n"))
			continue
		}
		displayNames[displayName] = true
		mutex.Unlock()
		name = displayName
		break
	}

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
	delete(displayNames, name)
	mutex.Unlock()
	broadcast <- fmt.Sprintf("\033[33m%s has left the chat\033[0m\n", name)
	conn.Close()
}

// handleRegisterCommand handles user registration
func handleRegisterCommand(conn net.Conn, message string) string {
	// Get client IP
	ip := conn.RemoteAddr().String()

	// Check rate limiting
	if isRateLimited(ip) {
		conn.Write([]byte("\033[1;31mToo many registration attempts. Please try again later.\033[0m\n"))
		return ""
	}

	parts := strings.SplitN(message, " ", 3)
	if len(parts) != 3 {
		conn.Write([]byte("Usage: /register <username> <password>\n"))
		return ""
	}
	username := strings.TrimSpace(parts[1])
	password := strings.TrimSpace(parts[2])

	// Validate username and password length
	if len(username) > 10 {
		conn.Write([]byte("\033[1;31mUsername must be 10 characters or less.\033[0m\n"))
		return ""
	}
	if len(password) > 10 {
		conn.Write([]byte("\033[1;31mPassword must be 10 characters or less.\033[0m\n"))
		return ""
	}

	// Check if username already exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		conn.Write([]byte("\033[1;31mError checking username. Please try again.\033[0m\n"))
		return ""
	}
	if count > 0 {
		conn.Write([]byte("\033[1;31mUsername already exists. Please choose another.\033[0m\n"))
		return ""
	}

	// Save user to database
	if err := saveUser(username, password); err != nil {
		conn.Write([]byte("\033[1;31mError registering user. Please try again.\033[0m\n"))
		return ""
	}

	conn.Write([]byte(fmt.Sprintf("\033[1;32mWelcome, %s! You can now start chatting.\033[0m\n", username)))
	return username
}

// handleLoginCommand handles user login
func handleLoginCommand(conn net.Conn, message string) string {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) != 3 {
		conn.Write([]byte("Usage: /login <username> <password>\n"))
		return ""
	}
	username := strings.TrimSpace(parts[1])
	password := strings.TrimSpace(parts[2])

	// Validate username and password length
	if len(username) > 10 {
		conn.Write([]byte("\033[1;31mUsername must be 10 characters or less.\033[0m\n"))
		return ""
	}
	if len(password) > 10 {
		conn.Write([]byte("\033[1;31mPassword must be 10 characters or less.\033[0m\n"))
		return ""
	}

	if !verifyUser(username, password) {
		conn.Write([]byte("\033[1;31mInvalid username or password.\033[0m\n"))
		return ""
	}

	conn.Write([]byte(fmt.Sprintf("\033[1;32mWelcome back, %s!\033[0m\n", username)))
	return username
}

// handleStatusCommand handles the /status command
func handleStatusCommand(conn net.Conn, message string) {
	parts := strings.SplitN(message, " ", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		conn.Write([]byte("\033[1;31mUsage: /status <set status>\033[0m\n"))
		return
	}
	newStatus := parts[1]
	mutex.Lock()
	username := clients[conn]
	mutex.Unlock()

	if err := updateUserStatus(username, newStatus); err != nil {
		conn.Write([]byte("\033[1;31mError updating status. Please try again.\033[0m\n"))
		return
	}

	conn.Write([]byte(fmt.Sprintf("\033[1;32mYour status has been set to: %s\033[0m\n", newStatus)))
}

// handleUsersCommand handles the /users command
func handleUsersCommand(conn net.Conn) {
	users, err := getAllUsers()
	if err != nil {
		conn.Write([]byte("\033[1;31mError retrieving users list.\033[0m\n"))
		return
	}

	for username, status := range users {
		if status != "" {
			conn.Write([]byte(fmt.Sprintf("\033[90m%s (%s)\033[0m\n", username, status)))
		} else {
			conn.Write([]byte("\033[90m" + username + "\033[0m\n"))
		}
	}
}

// handleBroadcasting sends messages to all connected clients
func handleBroadcasting() {
	for message := range broadcast {
		mutex.Lock()
		for conn := range clients {
			conn.Write([]byte(message))
		}
		mutex.Unlock()
	}
}

// handleExitCommand handles the /exit command
func handleExitCommand(conn net.Conn) {
	mutex.Lock()
	name := clients[conn]
	delete(clients, conn)
	delete(nameToConn, name)
	mutex.Unlock()

	// Notify everyone that the user has left
	broadcast <- fmt.Sprintf("\033[33m%s has left the chat\033[0m\n", name)

	// Send goodbye message to the exiting user
	conn.Write([]byte("\033[1;32mGoodbye! Thanks for chatting.\033[0m\n"))

	// Close the connection
	conn.Close()
}

// handleHelpCommand displays all available commands and their descriptions
func handleHelpCommand(conn net.Conn) {
	helpMessage := "\033[1;36mAvailable Commands:\033[0m\n\n" +
		"\033[1;33m/register <username> <password>\033[0m\n" +
		"    Register a new user account\n\n" +
		"\033[1;33m/login <username> <password>\033[0m\n" +
		"    Login to your account\n\n" +
		"\033[1;33m/users\033[0m\n" +
		"    List all currently connected users\n\n" +
		"\033[1;33m/private <username> <message>\033[0m\n" +
		"    Send a private message to a specific user\n\n" +
		"\033[1;33m/reply <message>\033[0m\n" +
		"    Reply to the last private message you received\n\n" +
		"\033[1;33m/exit\033[0m\n" +
		"    Exit the chat server\n\n" +
		"\033[1;33m/help\033[0m\n" +
		"    Display this help message\n\n" +
		"\033[1;36mRegular Messages:\033[0m\n" +
		"    Type any message without a command to broadcast to all users\n"

	conn.Write([]byte(helpMessage))
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
	// /exit command
	if strings.HasPrefix(message, "/exit") {
		handleExitCommand(conn)
		return true
	}
	// /help command
	if strings.HasPrefix(message, "/help") {
		handleHelpCommand(conn)
		return true
	}
	// /status command
	if strings.HasPrefix(message, "/status") {
		handleStatusCommand(conn, message)
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
		conn.Write([]byte("\033[1;31mNo private messages to reply to.\033[0m\n"))
		return
	}
	parts := strings.SplitN(message, " ", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		conn.Write([]byte("\033[1;31mUsage: /reply <message>\033[0m\n"))
		return
	}
	msg := parts[1]
	handlePrivateMessage(conn, fmt.Sprintf("/private %s %s", lastSender, msg))
}
