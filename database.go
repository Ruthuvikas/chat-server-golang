package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

// initDB initializes the database and creates necessary tables
func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./chat.db")
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	// Create users table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		password TEXT NOT NULL,
		status TEXT DEFAULT ''
	);
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}

	return nil
}

// saveUser saves a new user to the database
func saveUser(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, string(hashedPassword))
	return err
}

// verifyUser checks if the username and password match
func verifyUser(username, password string) bool {
	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&hashedPassword)
	if err != nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// updateUserStatus updates a user's status
func updateUserStatus(username, newStatus string) error {
	_, err := db.Exec("UPDATE users SET status = ? WHERE username = ?", newStatus, username)
	return err
}

// getUserStatus retrieves a user's status
func getUserStatus(username string) (string, error) {
	var status string
	err := db.QueryRow("SELECT status FROM users WHERE username = ?", username).Scan(&status)
	if err != nil {
		return "", err
	}
	return status, nil
}

// getAllUsers retrieves all users and their statuses
func getAllUsers() (map[string]string, error) {
	rows, err := db.Query("SELECT username, status FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make(map[string]string)
	for rows.Next() {
		var username, status string
		if err := rows.Scan(&username, &status); err != nil {
			return nil, err
		}
		users[username] = status
	}
	return users, nil
}

// closeDB closes the database connection
func closeDB() error {
	return db.Close()
}
