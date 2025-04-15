package main

import (
	"fmt"
	"net"
	"strings"
)

type PrivateMessage struct {
	sender    string
	recipient string
	message   string
}

var (
	privateMsg = make(chan PrivateMessage)
)

// HandlePrivateMessage processes private messages from clients
func handlePrivateMessage(conn net.Conn, message string) {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) != 3 {
		conn.Write([]byte("Usage: /private <username> <message>\n"))
		return
	}
	recipient := parts[1]
	content := parts[2]
	privateMsg <- PrivateMessage{
		sender:    clients[conn],
		recipient: recipient,
		message:   content,
	}
}

// HandlePrivateMessages processes the private message channel
func processPrivateMessages() {
	for msg := range privateMsg {
		mutex.Lock()
		conn, ok := nameToConn[msg.recipient]
		senderConn := nameToConn[msg.sender]
		mutex.Unlock()
		if ok {
			conn.Write([]byte(fmt.Sprintf("[Private from %s] %s\n", msg.sender, msg.message)))
		} else {
			senderConn.Write([]byte(fmt.Sprintf("User %s not found\n", msg.recipient)))
		}
	}
}
