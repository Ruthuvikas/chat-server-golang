package main

import (
	"bytes"
	"encoding/json"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
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

	// Test
	username := handleRegisterCommand(conn, message)

	// Verify
	if username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", username)
	}

	// Check if password was hashed and stored
	mutex.Lock()
	hashedPass, exists := nameToPass["testuser"]
	mutex.Unlock()
	if !exists {
		t.Error("User was not stored in nameToPass")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte("testpass")); err != nil {
		t.Error("Password was not properly hashed")
	}

	// Cleanup
	mutex.Lock()
	delete(nameToPass, "testuser")
	mutex.Unlock()
}

func TestHandleLoginCommand(t *testing.T) {
	// Setup
	conn, _ := createMockConn()
	// First register a user
	handleRegisterCommand(conn, "/register testuser testpass")

	// Test
	username := handleLoginCommand(conn, "/login testuser testpass")

	// Verify
	if username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", username)
	}

	// Cleanup
	mutex.Lock()
	delete(nameToPass, "testuser")
	mutex.Unlock()
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

	// Set up the client mapping
	mutex.Lock()
	clients[conn] = "testuser"
	nameToConn["testuser"] = conn
	mutex.Unlock()

	// Test
	handleStatusCommand(conn, "/status busy")

	// Give some time for the status to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify
	mutex.Lock()
	userStatus, exists := status["testuser"]
	mutex.Unlock()
	if !exists {
		t.Error("Status was not set")
	}
	if userStatus != "busy" {
		t.Errorf("Expected status 'busy', got '%s'", userStatus)
	}

	// Cleanup
	mutex.Lock()
	delete(clients, conn)
	delete(nameToConn, "testuser")
	delete(status, "testuser")
	mutex.Unlock()
}

func TestSaveUsersToFile(t *testing.T) {
	// Setup
	testUsers := map[string]string{
		"testuser1": "hashedpass1",
		"testuser2": "hashedpass2",
	}
	mutex.Lock()
	nameToPass = testUsers
	mutex.Unlock()

	// Test
	err := saveUsersToFile()
	if err != nil {
		t.Errorf("Error saving users to file: %v", err)
	}

	// Verify
	data, err := os.ReadFile("users.json")
	if err != nil {
		t.Errorf("Error reading users file: %v", err)
	}

	var loadedUsers map[string]string
	err = json.Unmarshal(data, &loadedUsers)
	if err != nil {
		t.Errorf("Error unmarshaling users: %v", err)
	}

	if len(loadedUsers) != len(testUsers) {
		t.Errorf("Expected %d users, got %d", len(testUsers), len(loadedUsers))
	}

	// Cleanup
	os.Remove("users.json")
	mutex.Lock()
	nameToPass = make(map[string]string)
	mutex.Unlock()
}
