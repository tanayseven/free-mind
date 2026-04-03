# Root Daemon Implementation Plan

## Overview

This document outlines the changes needed to update the root-daemon/main.go file to use the new IPC package instead of ZMQ.

## Current Implementation

The current implementation uses ZMQ for communication:
- Creates a ZMQ router on an available port
- Writes the port number to a file for the client to discover
- Listens for messages in a loop
- Processes messages and sends responses

## New Implementation

The new implementation will use our custom IPC package:
- Creates an IPC server using Unix sockets (Linux/Darwin) or named pipes (Windows)
- Writes the socket/pipe path to a file for the client to discover (optional, as the path is fixed)
- Listens for connections and handles them in a loop
- Processes messages and sends responses

## Code Changes

> **Status:** All code changes completed.

### Imports

Replace:
```go
import (
    "gopkg.in/zeromq/goczmq.v4"
)
```

With:
```go
import (
    "free-mind/ipc"
)
```

### Main Function

Replace:
```go
func main() {
    startPort := 5500
    endPort := 5600

    availablePort, err := FindAvailablePort(startPort, endPort)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    // Write the daemon port to the correct path
    portPath := FreeMindDaemonPortPath()
    err = os.WriteFile(portPath, []byte(fmt.Sprintf("%d", availablePort)), 0644)
    if err != nil {
        log.Printf("Error writing daemon port to file: %v", err)
        // Continue execution even if writing port fails
    } else {
        log.Printf("Daemon port %d written to %s", availablePort, portPath)
    }
    fmt.Printf("Found available port: %d\n", availablePort)
    router, err := goczmq.NewRouter(fmt.Sprintf("tcp://*:%d", availablePort))
    if err != nil {
        log.Fatal(err)
        return
    }
    defer router.Destroy()
    for {
        request, err := router.RecvMessage()
        if err != nil {
            log.Fatal(err)
        }
        // Parse the received JSON message
        var receivedMsg Message
        err = json.Unmarshal(request[1], &receivedMsg)
        if err != nil {
            log.Printf("Error unmarshaling JSON: %v", err)
            // Continue with error handling if needed
        }

        ProcessMessage(receivedMsg)

        log.Printf("router received action: '%s', content: '%s' from '%v'",
            receivedMsg.Action, receivedMsg.Content, request[0])

        // Create a response message
        responseMsg := Message{
            Action:  "response",
            Content: "Message received successfully",
        }

        // Marshal the response to JSON
        responseJSON, err := json.Marshal(responseMsg)
        if err != nil {
            log.Fatal(err)
        }

        // Send the response back
        err = router.SendFrame(request[0], goczmq.FlagMore)
        if err != nil {
            log.Fatal(err)
        }

        log.Printf("router sent response JSON")
        err = router.SendFrame(responseJSON, goczmq.FlagNone)
        if err != nil {
            log.Fatal(err)
        }
    }
}
```

With:
```go
func main() {
    // Create a new IPC server
    server := ipc.NewServer()
    
    // Write the socket/pipe path to the correct file (optional, as the path is fixed)
    socketPath := server.Path()
    socketPathFile := FreeMindDaemonSocketPath()
    err := os.WriteFile(socketPathFile, []byte(socketPath), 0644)
    if err != nil {
        log.Printf("Error writing socket path to file: %v", err)
        // Continue execution even if writing path fails
    } else {
        log.Printf("Daemon socket path %s written to %s", socketPath, socketPathFile)
    }
    
    // Start listening for connections
    err = server.Listen()
    if err != nil {
        log.Fatalf("Error starting IPC server: %v", err)
        return
    }
    defer server.Close()
    
    log.Printf("IPC server listening on %s", socketPath)
    
    // Accept connections in a loop
    for {
        // Accept a new connection
        conn, err := server.Accept()
        if err != nil {
            log.Printf("Error accepting connection: %v", err)
            continue
        }
        
        // Handle the connection in a goroutine
        go handleConnection(conn)
    }
}

func handleConnection(conn ipc.Connection) {
    defer conn.Close()
    
    // Receive the message
    receivedMsg, err := conn.Receive()
    if err != nil {
        log.Printf("Error receiving message: %v", err)
        return
    }
    
    // Process the message
    ProcessMessage(receivedMsg)
    
    log.Printf("Received action: '%s', content: '%s'", receivedMsg.Action, receivedMsg.Content)
    
    // Create a response message
    responseMsg := &ipc.Message{
        Action:  "response",
        Content: "Message received successfully",
    }
    
    // Send the response back
    err = conn.Send(responseMsg)
    if err != nil {
        log.Printf("Error sending response: %v", err)
        return
    }
    
    log.Printf("Sent response")
}
```

### FreeMindDaemonSocketPath Function

Add a new function to get the path to the socket/pipe path file:

```go
func FreeMindDaemonSocketPath() string {
    switch runtime.GOOS {
    case "windows":
        return `C:\Windows\System32\drivers\etc\free-mind-daemon-socket-path`
    case "darwin", "linux":
        return "/etc/free-mind-daemon-socket-path"
    default:
        return ""
    }
}
```

## Message Structure

The Message struct will be moved to the IPC package, so we can remove it from the root-daemon/main.go file.

## Testing

To test the new implementation:
- [x] Start the daemon
- [x] Use the client to connect to the daemon
- [x] Send messages and verify responses (unit tests in `root-daemon/daemon_test.go`)
- [ ] Test on different platforms (Linux, Darwin, Windows)