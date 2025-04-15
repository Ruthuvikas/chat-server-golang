//creating a tcp server that can handle multiple clients

package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

var (
	clients   = make(map[net.Conn]string) //map of clients
	broadcast = make(chan string)         //channel for broadcasting messages
	mutex     = &sync.Mutex{}             //mutex for thread safety
)

func main() {
	ln, err := net.Listen("tcp", ":8080") //listen on port 8080
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer ln.Close()

	go handleBroadcasting()

	fmt.Println("Server is running on port 8080")

	for {
		conn, err := ln.Accept() //accept incoming connections
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}
		go handleClient(conn) //handle client in a new goroutine
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
	clients[conn] = name
	broadcast <- fmt.Sprintf("%s has joined the chat\n", name)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		message = strings.TrimSpace(message)
		broadcast <- fmt.Sprintf("%s: %s\n", name, message)
	}
	mutex.Lock()
	delete(clients, conn)
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
