package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	NOT_FOUND = "HTTP/1.1 404 Not Found\r\n\r\n"
	OK        = "HTTP/1.1 200 OK\r\n\r\n"
)

func handleConnection(conn net.Conn) {
	var buf []byte = make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from connection", err.Error())
	}
	if n > 1024 {
		fmt.Println("Buffer overflow")
		os.Exit(1)
	}
	url := strings.Fields(string(buf))[1]
	if url == "/" {
		conn.Write([]byte(OK))
	} else {
		conn.Write([]byte(NOT_FOUND))
	}
	conn.Close()
}

func main() {
	// Uncomment this block to pass the first stage
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	handleConnection(conn)
}
