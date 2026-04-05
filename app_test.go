//go:build linux || darwin
// +build linux darwin

package main

import (
	"context"
	"sync"
	"testing"

	"free-mind/ipc"
)

// startMockDaemon spins up a Unix IPC server at a temp socket path and registers
// a cleanup that shuts it down.  The handler function is called in a goroutine
// for every incoming connection.  The function returns the socket path that was
// chosen so callers can point ipc.OverrideSocketPath at it.
func startMockDaemon(t *testing.T, handler func(ipc.Connection)) string {
	t.Helper()

	path := t.TempDir() + "/mock-daemon.sock"
	srv := ipc.NewServerAt(path)
	if err := srv.Listen(); err != nil {
		t.Fatalf("mock daemon Listen() error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := srv.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					return
				}
			}
			wg.Add(1)
			go func(c ipc.Connection) {
				defer wg.Done()
				defer c.Close()
				handler(c)
			}(conn)
		}
	}()

	t.Cleanup(func() {
		cancel()
		srv.Close()
		wg.Wait()
		ipc.OverrideSocketPath = ""
	})

	ipc.OverrideSocketPath = path
	return path
}

// echoHandler receives one message and replies with a standard success response.
func echoHandler(conn ipc.Connection) {
	msg, err := conn.Receive()
	if err != nil {
		return
	}
	_ = msg
	conn.Send(&ipc.Message{Action: "response", Content: "Message received successfully"}) //nolint:errcheck
}

func TestConnectToDaemon(t *testing.T) {
	startMockDaemon(t, func(conn ipc.Connection) {
		// Just accept and close — ConnectToDaemon only tests that a connection succeeds.
		conn.Close()
	})

	app := NewApp()
	result := app.ConnectToDaemon()

	if result == "" {
		t.Error("ConnectToDaemon() returned empty string")
	}
	// Should report success, not an error prefix.
	if len(result) >= 6 && result[:6] == "Error:" {
		t.Errorf("ConnectToDaemon() returned error: %s", result)
	}
}

func TestSendMessage_Success(t *testing.T) {
	startMockDaemon(t, echoHandler)

	msg := &ipc.Message{Action: "start", Content: ""}
	reply, err := sendMessage(msg)
	if err != nil {
		t.Fatalf("sendMessage() error: %v", err)
	}
	if reply.Action != "response" {
		t.Errorf("reply.Action = %q, want %q", reply.Action, "response")
	}
	if reply.Content != "Message received successfully" {
		t.Errorf("reply.Content = %q", reply.Content)
	}
}

func TestSendMessage_NoServer(t *testing.T) {
	// Point to a path where no server is listening.
	ipc.OverrideSocketPath = t.TempDir() + "/nonexistent.sock"
	t.Cleanup(func() { ipc.OverrideSocketPath = "" })

	_, err := sendMessage(&ipc.Message{Action: "start"})
	if err == nil {
		t.Error("sendMessage() should have returned an error when no server is listening")
	}
}

func TestAppSendBlockList(t *testing.T) {
	var receivedMsg *ipc.Message
	startMockDaemon(t, func(conn ipc.Connection) {
		msg, err := conn.Receive()
		if err == nil {
			receivedMsg = msg
		}
		conn.Send(&ipc.Message{Action: "response", Content: "Message received successfully"}) //nolint:errcheck
	})

	app := NewApp()
	ok := app.SendBlockList("reddit.com,twitter.com")
	if !ok {
		t.Error("SendBlockList() returned false")
	}
	if receivedMsg == nil {
		t.Fatal("mock daemon never received a message")
	}
	if receivedMsg.Action != "update" {
		t.Errorf("message action = %q, want %q", receivedMsg.Action, "update")
	}
	if receivedMsg.Content != "reddit.com,twitter.com" {
		t.Errorf("message content = %q, want %q", receivedMsg.Content, "reddit.com,twitter.com")
	}
}

func TestAppStartBlocking(t *testing.T) {
	var receivedAction string
	startMockDaemon(t, func(conn ipc.Connection) {
		msg, err := conn.Receive()
		if err == nil {
			receivedAction = msg.Action
		}
		conn.Send(&ipc.Message{Action: "response", Content: "Message received successfully"}) //nolint:errcheck
	})

	app := NewApp()
	ok := app.StartBlocking()
	if !ok {
		t.Error("StartBlocking() returned false")
	}
	if receivedAction != "start" {
		t.Errorf("daemon received action %q, want %q", receivedAction, "start")
	}
}

func TestAppStopBlocking(t *testing.T) {
	var receivedAction string
	startMockDaemon(t, func(conn ipc.Connection) {
		msg, err := conn.Receive()
		if err == nil {
			receivedAction = msg.Action
		}
		conn.Send(&ipc.Message{Action: "response", Content: "Message received successfully"}) //nolint:errcheck
	})

	app := NewApp()
	result := app.StopBlocking()
	if result != "Blocking stopped successfully" {
		t.Errorf("StopBlocking() = %q, want %q", result, "Blocking stopped successfully")
	}
	if receivedAction != "stop" {
		t.Errorf("daemon received action %q, want %q", receivedAction, "stop")
	}
}

func TestAppSendBlockList_Failure(t *testing.T) {
	// No server running — all IPC calls should fail gracefully.
	ipc.OverrideSocketPath = t.TempDir() + "/nonexistent.sock"
	t.Cleanup(func() { ipc.OverrideSocketPath = "" })

	app := NewApp()
	if app.SendBlockList("reddit.com") {
		t.Error("SendBlockList() should return false when the daemon is unreachable")
	}
	if app.StartBlocking() {
		t.Error("StartBlocking() should return false when the daemon is unreachable")
	}
	result := app.StopBlocking()
	if len(result) < 5 || result[:5] != "Error" {
		t.Errorf("StopBlocking() should return error string, got %q", result)
	}
}
