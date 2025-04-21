# Real-time Chat Server in Go

A high-performance, real-time chat server built with Go (Golang) that enables seamless communication between multiple clients using TCP protocol.

## Features

- Real-time message broadcasting
- Private messaging between users
- User registration and authentication
- Reply to the last private message sender with `/reply <message>`
- List all connected users with `/users` (including their status)
- Set your status with `/status`
- Exit chat gracefully with `/exit`
- Get help with all commands using `/help`
- Color-coded messages for better readability
- TCP-based communication
- Concurrent client handling
- Simple and efficient architecture
- Easy to deploy and scale
- SQLite database for persistent user storage
- Rate limiting for registration (3 attempts per minute)
- Unique display names enforcement
- Username and password length restrictions (max 10 characters)

## Security Features

- Password hashing using bcrypt
- Rate limiting for registration attempts
- SQL injection prevention
- Thread-safe operations
- Unique display name enforcement
- Input validation and sanitization

## Testing

The chat server includes comprehensive unit tests to ensure reliability and functionality. The test suite (`main_test.go`) covers:

- User registration and authentication
- Private messaging functionality
- Status updates and management
- User data persistence
- Network connection handling
- Rate limiting functionality
- Display name uniqueness
- Input validation

### Test Coverage

The test suite verifies:
- User registration with proper password hashing
- Login functionality with correct credentials
- Private message routing and delivery
- Status update and retrieval
- User data saving and loading
- Concurrent access handling
- Network connection management
- Rate limiting enforcement
- Display name uniqueness
- Input validation

### Running Tests

To run the tests:
```bash
make test
```

Or directly:
```bash
go test -v
```

## Video Demo

[![Chat Server Demo](https://img.youtube.com/vi/5UEvIQLwuIY/0.jpg)](https://www.youtube.com/watch?v=5UEvIQLwuIY)

## Prerequisites

- Go 1.16 or higher
- Basic understanding of TCP protocol
- Make (for using Makefile commands)
- SQLite3 (for database operations)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/chat-server-golang.git
cd chat-server-golang
```

2. Install dependencies:
```bash
make deps
```

3. Build the server:
```bash
make build
```

4. Run the server:
```bash
make run
```

Alternatively, you can run the server directly:
```bash
go run main.go database.go private_message.go
```

## Usage

The chat server runs on port `8080` by default. Connect to it using any TCP client:

```
telnet localhost 8080
```

### Commands

- To register a new account:
  ```
  /register <username> <password>
  ```
  - Username and password must be 10 characters or less
  - Maximum 3 registration attempts per minute per IP

- To login to your account:
  ```
  /login <username> <password>
  ```

- To set your display name:
  ```
  (Prompted after login/registration)
  ```
  - Display names must be unique
  - Case-sensitive

- To send a private message:
  ```
  /private <username> <message>
  ```

- To reply to the last private message sender:
  ```
  /reply <message>
  ```

- To list all connected users:
  ```
  /users
  ```

- To set your status:
  ```
  /status <your status message>
  ```

- To exit the chat server:
  ```
  /exit
  ```

- To view all available commands:
  ```
  /help
  ```

### Additional Make Commands

- `make build` - Build the chat server binary
- `make run` - Run the chat server
- `make clean` - Remove build artifacts
- `make test` - Run tests
- `make deps` - Install dependencies
- `make help` - Show all available commands

## Tutorial

For a detailed walkthrough of this project, check out my YouTube tutorial:
[Real-time Chat Server in Go - Tutorial](https://www.youtube.com/watch?v=5UEvIQLwuIY)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
