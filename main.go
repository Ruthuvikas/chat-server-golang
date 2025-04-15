//chat server using go

package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

var (
	clients    = make(map[net.Conn]string)
	nameToConn = make(map[string]net.Conn)
	broadcast  = make(chan string)
	mutex      = &sync.Mutex{}
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer ln.Close()
	go handleBroadcasting()
	go processPrivateMessages()

	fmt.Println("Server is running on port 8080")
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	conn.Write([]byte("Enter your name: "))
	reader := bufio.NewReader(conn)
	name, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading name:", err)
		return
	}
	name = strings.TrimSpace(name)
	mutex.Lock()
	clients[conn] = name
	nameToConn[name] = conn
	mutex.Unlock()
	broadcast <- fmt.Sprintf("%s has joined the chat\n", name)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		message = strings.TrimSpace(message)
		if strings.HasPrefix(message, "/private") {
			handlePrivateMessage(conn, message)
			continue
		}
		broadcast <- fmt.Sprintf("%s: %s\n", name, message)
	}
	mutex.Lock()
	delete(clients, conn)
	delete(nameToConn, name)
	mutex.Unlock()
	broadcast <- fmt.Sprintf("%s has left the chat\n", name)
	conn.Close()
}

func handleBroadcasting() {
	for message := range broadcast {
		mutex.Lock()
		for conn, _ := range clients {
			conn.Write([]byte(message))
		}
		mutex.Unlock()
	}
}
