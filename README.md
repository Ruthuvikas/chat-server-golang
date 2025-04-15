# Real-time Chat Server in Go

A high-performance, real-time chat server built with Go (Golang) that enables seamless communication between multiple clients.

## Features

- Real-time message broadcasting
- WebSocket-based communication
- Concurrent client handling
- Simple and efficient architecture
- Easy to deploy and scale

## Prerequisites

- Go 1.16 or higher
- Basic understanding of WebSocket protocol

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/chat-server-golang.git
cd chat-server-golang
```

2. Install dependencies:
```bash
go mod download
```

3. Run the server:
```bash
go run main.go
```

## Usage

The chat server runs on port `8080` by default. Connect to it using any WebSocket client:

```
ws://localhost:8080/ws
```

## Tutorial

For a detailed walkthrough of this project, check out my YouTube tutorial:
[Real-time Chat Server in Go - Tutorial](https://www.youtube.com/watch?v=5UEvIQLwuIY)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
