//go:build linux || darwin
// +build linux darwin

package ipc

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	// UnixSocketPath is the path to the Unix socket
	// Using /tmp for the socket path as it's typically writable by all users
	UnixSocketPath = "/tmp/tech.tanay.free-mind.sock"
)

// OverrideSocketPath redirects all NewServer/NewClient calls to the given path
// when non-empty. Intended for use in tests only.
var OverrideSocketPath string

// UnixServer implements the Server interface for Unix sockets
type UnixServer struct {
	listener *net.UnixListener
	path     string
}

// UnixConnection implements the Connection interface for Unix sockets
type UnixConnection struct {
	conn *net.UnixConn
}

// UnixClient implements the Client interface for Unix sockets
type UnixClient struct {
	conn *net.UnixConn
	path string
}

// NewServerAt returns a Server bound to path. Intended for use in tests.
func NewServerAt(path string) Server {
	return &UnixServer{path: path}
}

func init() {
	newUnixServer = func() Server {
		path := UnixSocketPath
		if OverrideSocketPath != "" {
			path = OverrideSocketPath
		}
		return &UnixServer{path: path}
	}

	newUnixClient = func() Client {
		path := UnixSocketPath
		if OverrideSocketPath != "" {
			path = OverrideSocketPath
		}
		return &UnixClient{path: path}
	}
}

// Listen implements the Server interface
func (s *UnixServer) Listen() error {
	log.Printf("DEBUG: UnixServer.Listen() called with path: %s", s.path)
	
	// Remove the socket file if it already exists
	if _, statErr := os.Stat(s.path); statErr == nil {
		log.Printf("DEBUG: Socket file already exists, attempting to remove it")
		if err := os.Remove(s.path); err != nil {
			log.Printf("Warning: Failed to remove existing socket file: %v", err)
		} else {
			log.Printf("DEBUG: Successfully removed existing socket file")
		}
	} else if os.IsNotExist(statErr) {
		log.Printf("DEBUG: Socket file does not exist, no need to remove")
	} else {
		log.Printf("DEBUG: Error checking socket file existence: %v", statErr)
	}
	
	// Create the socket directory if it doesn't exist
	dir := "/tmp"
	log.Printf("DEBUG: Checking socket directory: %s", dir)
	dirInfo, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Socket directory %s does not exist, attempting to create it", dir)
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Printf("Failed to create socket directory: %v", err)
				return fmt.Errorf("failed to create socket directory: %v", err)
			}
			log.Printf("Created socket directory %s", dir)
		} else {
			log.Printf("Error checking socket directory: %v", err)
			return fmt.Errorf("error checking socket directory: %v", err)
		}
	} else if !dirInfo.IsDir() {
		log.Printf("Socket path %s exists but is not a directory", dir)
		return fmt.Errorf("socket path %s exists but is not a directory", dir)
	} else {
		log.Printf("DEBUG: Socket directory %s exists and is a directory", dir)
	}
	
	// Check if we have write permission to the directory
	testFile := dir + "/test-write-permission"
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		log.Printf("No write permission to socket directory: %v", err)
		return fmt.Errorf("no write permission to socket directory: %v", err)
	} else {
		os.Remove(testFile)
		log.Printf("Socket directory exists and is writable")
	}
	
	// Create the Unix socket address
	log.Printf("DEBUG: Resolving Unix address for path: %s", s.path)
	addr, err := net.ResolveUnixAddr("unix", s.path)
	if err != nil {
		log.Printf("Failed to resolve Unix address: %v", err)
		return fmt.Errorf("failed to resolve Unix address: %v", err)
	}
	log.Printf("DEBUG: Successfully resolved Unix address")
	
	// Create the listener
	log.Printf("DEBUG: Creating Unix listener")
	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		log.Printf("Failed to create Unix listener: %v", err)
		return fmt.Errorf("failed to create Unix listener: %v", err)
	}
	log.Printf("DEBUG: Successfully created Unix listener")
	
	// Verify the socket file was created
	if _, statErr := os.Stat(s.path); statErr != nil {
		log.Printf("DEBUG: After ListenUnix - Socket file check error: %v", statErr)
	} else {
		log.Printf("DEBUG: After ListenUnix - Socket file exists at %s", s.path)
	}
	
	// Set permissions on the socket file
	log.Printf("DEBUG: Setting socket file permissions to 0666")
	if err := os.Chmod(s.path, 0666); err != nil {
		log.Printf("Failed to set socket file permissions: %v", err)
		listener.Close()
		return fmt.Errorf("failed to set socket file permissions: %v", err)
	}
	
	// Verify permissions were set correctly
	if fileInfo, statErr := os.Stat(s.path); statErr != nil {
		log.Printf("DEBUG: After chmod - Error checking socket file: %v", statErr)
	} else {
		log.Printf("DEBUG: After chmod - Socket file mode: %v", fileInfo.Mode())
	}
	
	log.Printf("Successfully created Unix socket at %s with permissions 0666", s.path)
	
	s.listener = listener
	return nil
}

// Accept implements the Server interface
func (s *UnixServer) Accept() (Connection, error) {
	conn, err := s.listener.AcceptUnix()
	if err != nil {
		return nil, err
	}
	
	return &UnixConnection{conn: conn}, nil
}

// Close implements the Server interface
func (s *UnixServer) Close() error {
	if s.listener != nil {
		err := s.listener.Close()
		os.Remove(s.path)
		return err
	}
	return nil
}

// Path implements the Server interface
func (s *UnixServer) Path() string {
	return s.path
}

// Send implements the Connection interface
func (c *UnixConnection) Send(msg *Message) error {
	encoder := json.NewEncoder(c.conn)
	return encoder.Encode(msg)
}

// Receive implements the Connection interface
func (c *UnixConnection) Receive() (*Message, error) {
	decoder := json.NewDecoder(c.conn)
	msg := &Message{}
	err := decoder.Decode(msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// Close implements the Connection interface
func (c *UnixConnection) Close() error {
	return c.conn.Close()
}

// Connect implements the Client interface
func (c *UnixClient) Connect() error {
	log.Printf("Attempting to connect to Unix socket at %s", c.path)
	
	// Check if the socket file exists
	fileInfo, err := os.Stat(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Socket file %s does not exist", c.path)
			// Check if the directory exists
			dir := "/tmp"
			dirInfo, dirErr := os.Stat(dir)
			if dirErr != nil {
				log.Printf("DEBUG: Error checking socket directory %s: %v", dir, dirErr)
			} else {
				log.Printf("DEBUG: Socket directory %s exists: isDir=%v, mode=%v", dir, dirInfo.IsDir(), dirInfo.Mode())
			}
		} else {
			log.Printf("Error checking socket file: %v", err)
		}
		return fmt.Errorf("socket file error: %v", err)
	} else {
		log.Printf("DEBUG: Socket file exists: isDir=%v, mode=%v, size=%v", fileInfo.IsDir(), fileInfo.Mode(), fileInfo.Size())
	}
	
	addr, err := net.ResolveUnixAddr("unix", c.path)
	if err != nil {
		log.Printf("Failed to resolve Unix address: %v", err)
		return fmt.Errorf("failed to resolve Unix address: %v", err)
	}
	
	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		log.Printf("Failed to connect to Unix socket: %v", err)
		return fmt.Errorf("failed to connect to Unix socket: %v", err)
	}
	
	log.Printf("Successfully connected to Unix socket at %s", c.path)
	c.conn = conn
	return nil
}

// Send implements the Client interface
func (c *UnixClient) Send(msg *Message) error {
	encoder := json.NewEncoder(c.conn)
	return encoder.Encode(msg)
}

// Receive implements the Client interface
func (c *UnixClient) Receive() (*Message, error) {
	decoder := json.NewDecoder(c.conn)
	msg := &Message{}
	err := decoder.Decode(msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// Close implements the Client interface
func (c *UnixClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Path implements the Client interface
func (c *UnixClient) Path() string {
	return c.path
}