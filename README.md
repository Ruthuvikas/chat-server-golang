# Real-time Chat Server in Go

A high-performance, real-time chat server built with Go (Golang) that enables seamless communication between multiple clients using TCP protocol.

## Features

- Real-time message broadcasting
- Private messaging between users
- User registration and authentication
- Reply to the last private message sender with `/reply <message>`
- List all connected users with `/users`
- Exit chat gracefully with `/exit`
- Get help with all commands using `/help`
- Color-coded messages for better readability
- TCP-based communication
- Concurrent client handling
- Simple and efficient architecture
- Easy to deploy and scale

## Prerequisites

- Go 1.16 or higher
- Basic understanding of TCP protocol
- Make (for using Makefile commands)

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
go run main.go private_message.go
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
- To login to your account:
  ```
  /login <username> <password>
  ```
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
