package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// Create a struct to hold the fields of the HTTP request
type Request struct {
	Method string
	URL    string
	Header map[string]string
	Body   []byte
}

// Create a struct to hold the fields of the HTTP response
type Response struct {
	Status string
	Header map[string]string
	Body   []byte
}

// Create a function to parse the HTTP request
func ParseRequest(data []byte) *Request {
	request := new(Request)
	lines := strings.Split(string(data), "\r\n")
	request.Method = strings.Split(lines[0], " ")[0]
	request.URL = strings.Split(lines[0], " ")[1]
	request.Header = make(map[string]string)
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			request.Body = []byte(strings.Join(lines[i+1:], "\r\n"))
			break
		}
		parts := strings.Split(lines[i], ": ")
		request.Header[parts[0]] = parts[1]
	}
	return request
}

// Create a function to serialize the HTTP response
func (response *Response) Serialize() []byte {
	data := fmt.Sprintf("HTTP/1.1 %s\r\n", response.Status)
	for key, value := range response.Header {
		data += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	data += "\r\n"
	data += string(response.Body)
	return []byte(data)
}

func HandleRequest(request *Request) *Response {
	response := new(Response)
	response.Header = make(map[string]string)
	switch url := request.URL; {
	case url == "/":
		response.Status = "200 OK"
		return response
	case strings.HasPrefix(url, "/echo/"):
		response.Status = "200 OK"
		response.Body = []byte(strings.TrimLeft(url, "/echo/"))
		response.Header["Content-Type"] = "text/plain"
		response.Header["Content-Length"] = fmt.Sprintf("%d", len(response.Body))
		return response
	case url == "/user-agent":
		response.Status = "200 OK"
		response.Body = []byte(request.Header["User-Agent"])
		response.Header["Content-Type"] = "text/plain"
		response.Header["Content-Length"] = fmt.Sprintf("%d", len(response.Body))
		return response
	case strings.HasPrefix(url, "/files/"):
		response = HandleFileRequest(request)
		fmt.Println(response)
		return response
	default:
		response.Status = "404 Not Found"
		return response
	}
}

func HandleFileRequest(request *Request) *Response {
	response := new(Response)
	response.Header = make(map[string]string)
	filename := strings.TrimPrefix(request.URL, "/files/")
	dir := os.Args[2]
	if request.Method == "GET" {
		data, err := os.ReadFile(filepath.Join(dir, filename))
		if err != nil {
			response.Status = "404 Not Found"
			return response
		}
		response.Status = "200 OK"
		response.Body = data
		response.Header["Content-Type"] = "application/octet-stream"
		response.Header["Content-Length"] = fmt.Sprintf("%d", len(data))
	} else if request.Method == "POST" {
		f, err := os.Create(filepath.Join(dir, filename))
		if err != nil {
			response.Status = "500 Internal Server Error"
			return response
		}
		defer f.Close()
		writer := bufio.NewWriter(f)
		_, err = writer.Write(request.Body)
		writer.Flush()
		if err != nil {
			response.Status = "500 Internal Server Error"
			return response
		}
		response.Status = "201 Created"
		fmt.Println("Created file", filename)
	}
	return response
}

// Create a function to handle the connection
func HandleConnection(conn net.Conn) {
	var buf []byte = make([]byte, 8192)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from connection", err.Error())
	}
	if n > 8192 {
		fmt.Println("Buffer overflow")
		os.Exit(1)
	}
	request := ParseRequest(buf)
	response := HandleRequest(request)
	fmt.Println(string(response.Serialize()))
	conn.Write(response.Serialize())
}

func main() {
	// Uncomment this block to pass the first stage
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go HandleConnection(conn)
	}
}
