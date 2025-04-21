package main

import (
	"bytes"
	"net"
	"strings"
	"testing"
	"time"
)

// Test helper function to create a mock connection
func createMockConn() (net.Conn, *bytes.Buffer) {
	client, server := net.Pipe()
	buf := bytes.NewBuffer(nil)
	go func() {
		for {
			data := make([]byte, 1024)
			n, err := server.Read(data)
			if err != nil {
				return
			}
			buf.Write(data[:n])
		}
	}()
	return client, buf
}

func TestHandleRegisterCommand(t *testing.T) {
	// Setup
	conn, _ := createMockConn()
	message := "/register testuser testpass"

	// Initialize database
	if err := initDB(); err != nil {
		t.Fatalf("Error initializing database: %v", err)
	}
	defer closeDB()

	// Test
	username := handleRegisterCommand(conn, message)

	// Verify
	if username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", username)
	}

	// Verify user was saved to database
	if !verifyUser("testuser", "testpass") {
		t.Error("User was not properly saved to database")
	}

	// Cleanup
	db.Exec("DELETE FROM users WHERE username = ?", "testuser")
}

func TestHandleLoginCommand(t *testing.T) {
	// Setup
	conn, _ := createMockConn()
	// Initialize database
	if err := initDB(); err != nil {
		t.Fatalf("Error initializing database: %v", err)
	}
	defer closeDB()

	// First register a user
	handleRegisterCommand(conn, "/register testuser testpass")

	// Test
	username := handleLoginCommand(conn, "/login testuser testpass")

	// Verify
	if username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", username)
	}

	// Cleanup
	db.Exec("DELETE FROM users WHERE username = ?", "testuser")
}

func TestHandlePrivateMessage(t *testing.T) {
	// Setup
	senderConn, _ := createMockConn()
	recipientConn, recipientBuf := createMockConn()

	// Start the private message processor in a goroutine
	go processPrivateMessages()

	// Register and login both users
	mutex.Lock()
	clients[senderConn] = "sender"
	clients[recipientConn] = "recipient"
	nameToConn["sender"] = senderConn
	nameToConn["recipient"] = recipientConn
	mutex.Unlock()

	// Test
	message := "/private recipient Hello!"
	handlePrivateMessage(senderConn, message)

	// Give some time for the message to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify
	mutex.Lock()
	_, exists := lastPrivateSender["recipient"]
	mutex.Unlock()
	if !exists {
		t.Error("Last private sender was not recorded")
	}

	// Check if recipient received the message
	recipientOutput := recipientBuf.String()
	expectedMessage := "[Private from sender] Hello!"
	if !strings.Contains(recipientOutput, expectedMessage) {
		t.Errorf("Recipient did not receive the message. Got: %s, Want: %s", recipientOutput, expectedMessage)
	}

	// Cleanup
	mutex.Lock()
	delete(clients, senderConn)
	delete(clients, recipientConn)
	delete(nameToConn, "sender")
	delete(nameToConn, "recipient")
	delete(lastPrivateSender, "recipient")
	mutex.Unlock()
}

func TestHandleStatusCommand(t *testing.T) {
	// Setup
	conn, _ := createMockConn()
	if err := initDB(); err != nil {
		t.Fatalf("Error initializing database: %v", err)
	}
	defer closeDB()

	// Register and login a user
	handleRegisterCommand(conn, "/register testuser testpass")
	mutex.Lock()
	clients[conn] = "testuser"
	nameToConn["testuser"] = conn
	mutex.Unlock()

	// Test
	handleStatusCommand(conn, "/status busy")

	// Verify
	status, err := getUserStatus("testuser")
	if err != nil {
		t.Error("Error getting user status:", err)
	}
	if status != "busy" {
		t.Errorf("Expected status 'busy', got '%s'", status)
	}

	// Cleanup
	mutex.Lock()
	delete(clients, conn)
	delete(nameToConn, "testuser")
	mutex.Unlock()
	db.Exec("DELETE FROM users WHERE username = ?", "testuser")
}

func TestSaveUsersToFile(t *testing.T) {
	// This test is no longer needed as we're using database storage now
	t.Skip("Skipping test as we're using database storage now")
}
