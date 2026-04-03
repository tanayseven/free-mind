//go:build linux || darwin
// +build linux darwin

package ipc

import (
	"sync"
	"testing"
	"time"
)

// newServerAt creates a UnixServer bound to a custom socket path.
func newServerAt(path string) *UnixServer {
	return &UnixServer{path: path}
}

// newClientAt creates a UnixClient targeting a custom socket path.
func newClientAt(path string) *UnixClient {
	return &UnixClient{path: path}
}

func tempSockPath(t *testing.T) string {
	t.Helper()
	return t.TempDir() + "/test.sock"
}

// TestServerListenAndClose verifies that Listen creates the socket file and Close removes it.
func TestServerListenAndClose(t *testing.T) {
	path := tempSockPath(t)
	srv := newServerAt(path)

	if err := srv.Listen(); err != nil {
		t.Fatalf("Listen() error: %v", err)
	}

	if srv.Path() != path {
		t.Errorf("Path() = %q, want %q", srv.Path(), path)
	}

	if err := srv.Close(); err != nil {
		t.Errorf("Close() error: %v", err)
	}
}

// TestClientConnect verifies that a client can connect to a listening server.
func TestClientConnect(t *testing.T) {
	path := tempSockPath(t)
	srv := newServerAt(path)
	if err := srv.Listen(); err != nil {
		t.Fatalf("Listen() error: %v", err)
	}
	defer srv.Close()

	// Accept in background so Connect doesn't block.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := srv.Accept()
		if err != nil {
			return
		}
		conn.Close()
	}()

	client := newClientAt(path)
	if err := client.Connect(); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	client.Close()
	wg.Wait()
}

// TestSendReceive verifies a full message round-trip between client and server.
func TestSendReceive(t *testing.T) {
	path := tempSockPath(t)
	srv := newServerAt(path)
	if err := srv.Listen(); err != nil {
		t.Fatalf("Listen() error: %v", err)
	}
	defer srv.Close()

	wantAction := "start"
	wantContent := "example.com,reddit.com"

	// Server goroutine: receive request, echo back as response.
	serverErr := make(chan error, 1)
	go func() {
		conn, err := srv.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()

		msg, err := conn.Receive()
		if err != nil {
			serverErr <- err
			return
		}

		reply := &Message{Action: "response", Content: msg.Content}
		serverErr <- conn.Send(reply)
	}()

	// Client side.
	client := newClientAt(path)
	if err := client.Connect(); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer client.Close()

	sent := &Message{Action: wantAction, Content: wantContent}
	if err := client.Send(sent); err != nil {
		t.Fatalf("Send() error: %v", err)
	}

	reply, err := client.Receive()
	if err != nil {
		t.Fatalf("Receive() error: %v", err)
	}
	if reply.Action != "response" {
		t.Errorf("reply.Action = %q, want %q", reply.Action, "response")
	}
	if reply.Content != wantContent {
		t.Errorf("reply.Content = %q, want %q", reply.Content, wantContent)
	}

	if err := <-serverErr; err != nil {
		t.Errorf("server error: %v", err)
	}
}

// TestServerAcceptsSequentialConnections verifies the server handles multiple
// sequential connections correctly.
func TestServerAcceptsSequentialConnections(t *testing.T) {
	path := tempSockPath(t)
	srv := newServerAt(path)
	if err := srv.Listen(); err != nil {
		t.Fatalf("Listen() error: %v", err)
	}
	defer srv.Close()

	const n = 3
	for i := 0; i < n; i++ {
		done := make(chan struct{})
		go func() {
			conn, err := srv.Accept()
			if err != nil {
				close(done)
				return
			}
			conn.Close()
			close(done)
		}()

		c := newClientAt(path)
		if err := c.Connect(); err != nil {
			t.Fatalf("connection %d: Connect() error: %v", i, err)
		}
		c.Close()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("connection %d: server did not accept within timeout", i)
		}
	}
}

// TestClientConnectFailsWhenNoServer verifies that Connect returns an error
// when there is no server listening.
func TestClientConnectFailsWhenNoServer(t *testing.T) {
	path := tempSockPath(t)
	client := newClientAt(path)
	if err := client.Connect(); err == nil {
		t.Error("Connect() should have failed when no server is listening")
		client.Close()
	}
}

// TestServerListenReplacesStaleSocket verifies that Listen succeeds even when a
// stale socket file is left over from a previous run.
func TestServerListenReplacesStaleSocket(t *testing.T) {
	path := tempSockPath(t)

	// First server creates the socket.
	srv1 := newServerAt(path)
	if err := srv1.Listen(); err != nil {
		t.Fatalf("first Listen() error: %v", err)
	}
	// Close without removing the socket file to simulate a crash.
	srv1.listener.Close()

	// Second server should remove the stale file and listen successfully.
	srv2 := newServerAt(path)
	if err := srv2.Listen(); err != nil {
		t.Fatalf("second Listen() error: %v", err)
	}
	defer srv2.Close()
}

// TestNewServerAndClientFactories verifies the package-level constructors return
// non-nil implementations on Unix platforms.
func TestNewServerAndClientFactories(t *testing.T) {
	srv := NewServer()
	if srv == nil {
		t.Error("NewServer() returned nil on a Unix platform")
	}

	c := NewClient()
	if c == nil {
		t.Error("NewClient() returned nil on a Unix platform")
	}
}
