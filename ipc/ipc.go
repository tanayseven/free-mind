// Package ipc provides an interface for inter-process communication
package ipc

import (
	"runtime"
)

// Message represents the JSON structure for communication
type Message struct {
	Action  string `json:"action"`
	Content string `json:"content"`
}

// Server defines the interface for an IPC server
type Server interface {
	// Listen starts the server and listens for connections
	Listen() error
	
	// Accept accepts a new connection
	Accept() (Connection, error)
	
	// Close closes the server
	Close() error
	
	// Path returns the server's socket/pipe path
	Path() string
}

// Connection defines the interface for an IPC connection
type Connection interface {
	// Send sends a message over the connection
	Send(msg *Message) error
	
	// Receive receives a message from the connection
	Receive() (*Message, error)
	
	// Close closes the connection
	Close() error
}

// Client defines the interface for an IPC client
type Client interface {
	// Connect connects to a server
	Connect() error
	
	// Send sends a message to the server
	Send(msg *Message) error
	
	// Receive receives a message from the server
	Receive() (*Message, error)
	
	// Close closes the client connection
	Close() error
	
	// Path returns the client's socket/pipe path
	Path() string
}

// These functions will be implemented in platform-specific files
var newUnixServer func() Server
var newWindowsServer func() Server
var newUnixClient func() Client
var newWindowsClient func() Client

// NewServer creates a new server based on the current platform
func NewServer() Server {
	switch runtime.GOOS {
	case "linux", "darwin":
		if newUnixServer != nil {
			return newUnixServer()
		}
	case "windows":
		if newWindowsServer != nil {
			return newWindowsServer()
		}
	}
	return nil
}

// NewClient creates a new client based on the current platform
func NewClient() Client {
	switch runtime.GOOS {
	case "linux", "darwin":
		if newUnixClient != nil {
			return newUnixClient()
		}
	case "windows":
		if newWindowsClient != nil {
			return newWindowsClient()
		}
	}
	return nil
}