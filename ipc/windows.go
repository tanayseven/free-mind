//go:build windows
// +build windows

package ipc

import (
	"encoding/json"
	"io"
	"time"

	"github.com/Microsoft/go-winio"
)

const (
	// WindowsPipePath is the path to the Windows named pipe
	WindowsPipePath = `\\.\pipe\tech.tanay.free-mind`
)

// WindowsServer implements the Server interface for Windows named pipes
type WindowsServer struct {
	listener *winio.PipeListener
	path     string
}

// WindowsConnection implements the Connection interface for Windows named pipes
type WindowsConnection struct {
	conn io.ReadWriteCloser
}

// WindowsClient implements the Client interface for Windows named pipes
type WindowsClient struct {
	conn io.ReadWriteCloser
	path string
}

func init() {
	newWindowsServer = func() Server {
		return &WindowsServer{
			path: WindowsPipePath,
		}
	}

	newWindowsClient = func() Client {
		return &WindowsClient{
			path: WindowsPipePath,
		}
	}
}

// Listen implements the Server interface
func (s *WindowsServer) Listen() error {
	// Configure the pipe
	config := &winio.PipeConfig{
		SecurityDescriptor: "",
		MessageMode:        true,
		InputBufferSize:    65536,
		OutputBufferSize:   65536,
	}
	
	// Create the listener
	listener, err := winio.ListenPipe(s.path, config)
	if err != nil {
		return err
	}
	
	s.listener = listener
	return nil
}

// Accept implements the Server interface
func (s *WindowsServer) Accept() (Connection, error) {
	conn, err := s.listener.Accept()
	if err != nil {
		return nil, err
	}
	
	return &WindowsConnection{conn: conn}, nil
}

// Close implements the Server interface
func (s *WindowsServer) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Path implements the Server interface
func (s *WindowsServer) Path() string {
	return s.path
}

// Send implements the Connection interface
func (c *WindowsConnection) Send(msg *Message) error {
	encoder := json.NewEncoder(c.conn)
	return encoder.Encode(msg)
}

// Receive implements the Connection interface
func (c *WindowsConnection) Receive() (*Message, error) {
	decoder := json.NewDecoder(c.conn)
	msg := &Message{}
	err := decoder.Decode(msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// Close implements the Connection interface
func (c *WindowsConnection) Close() error {
	return c.conn.Close()
}

// Connect implements the Client interface
func (c *WindowsClient) Connect() error {
	// Configure the timeout
	timeout := 5 * time.Second
	
	// Connect to the pipe
	conn, err := winio.DialPipe(c.path, &timeout)
	if err != nil {
		return err
	}
	
	c.conn = conn
	return nil
}

// Send implements the Client interface
func (c *WindowsClient) Send(msg *Message) error {
	encoder := json.NewEncoder(c.conn)
	return encoder.Encode(msg)
}

// Receive implements the Client interface
func (c *WindowsClient) Receive() (*Message, error) {
	decoder := json.NewDecoder(c.conn)
	msg := &Message{}
	err := decoder.Decode(msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// Close implements the Client interface
func (c *WindowsClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Path implements the Client interface
func (c *WindowsClient) Path() string {
	return c.path
}