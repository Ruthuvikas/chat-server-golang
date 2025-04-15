// Package main contains the private messaging functionality for the chat server
package main

import (
	"fmt"
	"net"
	"strings"
)

// PrivateMessage represents a private message between two users
type PrivateMessage struct {
	sender    string // Username of the sender
	recipient string // Username of the recipient
	message   string // The actual message content
}

// privateMsg channel for sending private messages between goroutines
var (
	privateMsg = make(chan PrivateMessage)
)

// handlePrivateMessage processes incoming private message requests
// Format: /private <username> <message>
func handlePrivateMessage(conn net.Conn, message string) {
	// Split the message into parts: command, recipient, and content
	parts := strings.SplitN(message, " ", 3)
	if len(parts) != 3 {
		conn.Write([]byte("Usage: /private <username> <message>\n"))
		return
	}

	// Extract recipient and message content
	recipient := parts[1]
	content := parts[2]

	// Create and send the private message
	privateMsg <- PrivateMessage{
		sender:    clients[conn],
		recipient: recipient,
		message:   content,
	}
}

// processPrivateMessages handles the private message channel
// It receives messages and delivers them to the intended recipient
func processPrivateMessages() {
	for msg := range privateMsg {
		mutex.Lock()
		// Look up the recipient's connection
		conn, ok := nameToConn[msg.recipient]
		// Get the sender's connection for error messages
		senderConn := nameToConn[msg.sender]
		mutex.Unlock()

		if ok {
			// Record the last private sender for the recipient
			mutex.Lock()
			lastPrivateSender[msg.recipient] = msg.sender
			mutex.Unlock()
			// Send the message to the recipient
			conn.Write([]byte(fmt.Sprintf("\033[34m[Private from %s] %s\033[0m\n", msg.sender, msg.message)))
		} else {
			// Notify sender if recipient is not found
			senderConn.Write([]byte(fmt.Sprintf("User %s not found\n", msg.recipient)))
		}
	}
}
