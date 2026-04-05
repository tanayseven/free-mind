# App Implementation Plan

## Overview

This document outlines the changes needed to update the app.go file to use the new IPC package instead of ZMQ.

## Current Implementation

The current implementation uses ZMQ for communication:
- Reads the daemon port from a file
- Creates a ZMQ dealer to connect to the daemon
- Sends messages to the daemon and receives responses

## New Implementation

The new implementation will use our custom IPC package:
- Creates an IPC client using Unix sockets (Linux/Darwin) or named pipes (Windows)
- Connects to the daemon using the fixed socket/pipe path
- Sends messages to the daemon and receives responses

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

### Global Variables

Replace:
```go
// Global ZMQ dealer
var dealer *goczmq.Sock
```

With:
```go
// Global IPC client
var client ipc.Client
```

### FetchDaemonPort Method

Replace:
```go
func (a *App) FetchDaemonPort() string {
    // Read the contents of the file from daemon port path
    portPath := FreeMindDaemonPortPath()
    portData, err := os.ReadFile(portPath)
    if err != nil {
        log.Printf("Error reading daemon port file: %v", err)
        return fmt.Sprintf("Error: %v", err)
    }

    // Convert port string to integer
    portStr := strings.TrimSpace(string(portData))
    port, err := strconv.Atoi(portStr)
    if err != nil {
        log.Printf("Error converting port to integer: %v", err)
        return fmt.Sprintf("Error: %v", err)
    }

    // Start a zmq dealer on that port
    var dealerErr error
    dealer, dealerErr = goczmq.NewDealer(fmt.Sprintf("tcp://127.0.0.1:%d", port))
    if dealerErr != nil {
        log.Printf("Error creating ZMQ dealer: %v", dealerErr)
        return fmt.Sprintf("Error: %v", dealerErr)
    }

    return fmt.Sprintf("Connected to daemon on port %d", port)
}
```

With:
```go
func (a *App) ConnectToDaemon() string {
    // Create a new IPC client
    client = ipc.NewClient()
    
    // Connect to the daemon
    err := client.Connect()
    if err != nil {
        log.Printf("Error connecting to daemon: %v", err)
        return fmt.Sprintf("Error: %v", err)
    }
    
    return fmt.Sprintf("Connected to daemon at %s", client.Path())
}
```

### SendBlockList Method

Replace:
```go
func (a *App) SendBlockList(list string) bool {
    // Check if the zmq dealer is set in the variable else exit with error
    if dealer == nil {
        log.Println("ZMQ dealer not initialized. Call FetchDaemonPort first.")
        return false
    }

    // Create a message to send
    msg := Message{
        Action:  "update",
        Content: list,
    }

    // Marshal the message to JSON
    jsonData, err := json.Marshal(msg)
    if err != nil {
        log.Printf("Error marshaling JSON: %v", err)
        return false
    }

    // Send the JSON data
    err = dealer.SendFrame(jsonData, goczmq.FlagNone)
    if err != nil {
        log.Printf("Error sending message: %v", err)
        return false
    }

    // Receive the reply
    reply, err := dealer.RecvMessage()
    if err != nil {
        log.Printf("Error receiving response: %v", err)
        return false
    }

    // Parse the received JSON response
    var responseMsg Message
    err = json.Unmarshal(reply[0], &responseMsg)
    if err != nil {
        log.Printf("Error unmarshaling response JSON: %v", err)
        return false
    }

    // Check if the daemon received the message successfully
    return responseMsg.Action == "response" && responseMsg.Content == "Message received successfully"
}
```

With:
```go
func (a *App) SendBlockList(list string) bool {
    // Check if the IPC client is initialized
    if client == nil {
        log.Println("IPC client not initialized. Call ConnectToDaemon first.")
        return false
    }
    
    // Create a message to send
    msg := &ipc.Message{
        Action:  "update",
        Content: list,
    }
    
    // Send the message
    err := client.Send(msg)
    if err != nil {
        log.Printf("Error sending message: %v", err)
        return false
    }
    
    // Receive the reply
    reply, err := client.Receive()
    if err != nil {
        log.Printf("Error receiving response: %v", err)
        return false
    }
    
    // Check if the daemon received the message successfully
    return reply.Action == "response" && reply.Content == "Message received successfully"
}
```

### StartBlocking Method

Replace:
```go
func (a *App) StartBlocking() bool {
    // Check if the zmq dealer is set in the variable else exit with error
    if dealer == nil {
        log.Println("ZMQ dealer not initialized. Call FetchDaemonPort first.")
        return false
    }

    // Create a message to send
    msg := Message{
        Action:  "start",
        Content: "",
    }

    // Marshal the message to JSON
    jsonData, err := json.Marshal(msg)
    if err != nil {
        log.Printf("Error marshaling JSON: %v", err)
        return false
    }

    // Send the JSON data
    err = dealer.SendFrame(jsonData, goczmq.FlagNone)
    if err != nil {
        log.Printf("Error sending message: %v", err)
        return false
    }

    // Receive the reply
    reply, err := dealer.RecvMessage()
    if err != nil {
        log.Printf("Error receiving response: %v", err)
        return false
    }

    // Parse the received JSON response
    var responseMsg Message
    err = json.Unmarshal(reply[0], &responseMsg)
    if err != nil {
        log.Printf("Error unmarshaling response JSON: %v", err)
        return false
    }

    // Check if the daemon received the message successfully
    return responseMsg.Action == "response" && responseMsg.Content == "Message received successfully"
}
```

With:
```go
func (a *App) StartBlocking() bool {
    // Check if the IPC client is initialized
    if client == nil {
        log.Println("IPC client not initialized. Call ConnectToDaemon first.")
        return false
    }
    
    // Create a message to send
    msg := &ipc.Message{
        Action:  "start",
        Content: "",
    }
    
    // Send the message
    err := client.Send(msg)
    if err != nil {
        log.Printf("Error sending message: %v", err)
        return false
    }
    
    // Receive the reply
    reply, err := client.Receive()
    if err != nil {
        log.Printf("Error receiving response: %v", err)
        return false
    }
    
    // Check if the daemon received the message successfully
    return reply.Action == "response" && reply.Content == "Message received successfully"
}
```

### StopBlocking Method

Replace:
```go
func (a *App) StopBlocking() string {
    // Check if the zmq dealer is set in the variable else exit with error
    if dealer == nil {
        return "Error: ZMQ dealer not initialized. Call FetchDaemonPort first."
    }

    // Create a message to send
    msg := Message{
        Action:  "stop",
        Content: "",
    }

    // Marshal the message to JSON
    jsonData, err := json.Marshal(msg)
    if err != nil {
        return fmt.Sprintf("Error marshaling JSON: %v", err)
    }

    // Send the JSON data
    err = dealer.SendFrame(jsonData, goczmq.FlagNone)
    if err != nil {
        return fmt.Sprintf("Error sending message: %v", err)
    }

    // Receive the reply
    reply, err := dealer.RecvMessage()
    if err != nil {
        return fmt.Sprintf("Error receiving response: %v", err)
    }

    // Parse the received JSON response
    var responseMsg Message
    err = json.Unmarshal(reply[0], &responseMsg)
    if err != nil {
        return fmt.Sprintf("Error unmarshaling response JSON: %v", err)
    }

    // Check if the daemon received the message successfully
    if responseMsg.Action == "response" && responseMsg.Content == "Message received successfully" {
        return "Blocking stopped successfully"
    } else {
        return "Failed to stop blocking"
    }
}
```

With:
```go
func (a *App) StopBlocking() string {
    // Check if the IPC client is initialized
    if client == nil {
        return "Error: IPC client not initialized. Call ConnectToDaemon first."
    }
    
    // Create a message to send
    msg := &ipc.Message{
        Action:  "stop",
        Content: "",
    }
    
    // Send the message
    err := client.Send(msg)
    if err != nil {
        return fmt.Sprintf("Error sending message: %v", err)
    }
    
    // Receive the reply
    reply, err := client.Receive()
    if err != nil {
        return fmt.Sprintf("Error receiving response: %v", err)
    }
    
    // Check if the daemon received the message successfully
    if reply.Action == "response" && reply.Content == "Message received successfully" {
        return "Blocking stopped successfully"
    } else {
        return "Failed to stop blocking"
    }
}
```

### CheckDaemonInstalled Method

Replace:
```go
func (a *App) CheckDaemonInstalled() bool {
    log.Println("Checking if daemon is installed and running...")

    // Check if the daemon binary exists
    destPath := a.GetDaemonBinaryDestination()
    if destPath == "" {
        log.Printf("Unsupported platform: %s", runtime.GOOS)
        return false
    }

    log.Printf("Checking for daemon binary at: %s", destPath)

    // Check if the binary file exists
    _, err := os.Stat(destPath)
    if err != nil {
        if os.IsNotExist(err) {
            log.Println("Daemon binary not found at:", destPath)
        } else {
            log.Printf("Error checking daemon binary: %v", err)
        }
        return false
    }
    log.Println("Daemon binary exists")

    // Check if the daemon port file exists
    portPath := FreeMindDaemonPortPath()
    log.Printf("Checking for daemon port file at: %s", portPath)

    _, err = os.Stat(portPath)
    if err != nil {
        if os.IsNotExist(err) {
            log.Println("Daemon port file not found at:", portPath)
        } else {
            log.Printf("Error checking daemon port file: %v", err)
        }
        return false
    }
    log.Println("Daemon port file exists")

    // Try to connect to the daemon
    portData, err := os.ReadFile(portPath)
    if err != nil {
        log.Printf("Error reading daemon port file: %v", err)
        return false
    }

    // Convert port string to integer
    portStr := strings.TrimSpace(string(portData))
    log.Printf("Daemon port from file: %s", portStr)

    port, err := strconv.Atoi(portStr)
    if err != nil {
        log.Printf("Error converting port to integer: %v", err)
        return false
    }

    // Try to create a ZMQ dealer to test the connection
    testDealer, err := goczmq.NewDealer(fmt.Sprintf("tcp://127.0.0.1:%d", port))
    if err != nil {
        log.Printf("Error connecting to daemon: %v", err)
        return false
    }
    testDealer.Destroy()

    log.Println("Daemon is installed and running")
    return true
}
```

With:
```go
func (a *App) CheckDaemonInstalled() bool {
    log.Println("Checking if daemon is installed and running...")
    
    // Check if the daemon binary exists
    destPath := a.GetDaemonBinaryDestination()
    if destPath == "" {
        log.Printf("Unsupported platform: %s", runtime.GOOS)
        return false
    }
    
    log.Printf("Checking for daemon binary at: %s", destPath)
    
    // Check if the binary file exists
    _, err := os.Stat(destPath)
    if err != nil {
        if os.IsNotExist(err) {
            log.Println("Daemon binary not found at:", destPath)
        } else {
            log.Printf("Error checking daemon binary: %v", err)
        }
        return false
    }
    log.Println("Daemon binary exists")
    
    // Try to connect to the daemon
    testClient := ipc.NewClient()
    err = testClient.Connect()
    if err != nil {
        log.Printf("Error connecting to daemon: %v", err)
        return false
    }
    testClient.Close()
    
    log.Println("Daemon is installed and running")
    return true
}
```

## Message Structure

The Message struct will be moved to the IPC package, so we can remove it from the app.go file.

## Testing

To test the new implementation:
- [x] Start the daemon
- [x] Use the app to connect to the daemon
- [x] Send messages and verify responses (unit tests in `app_test.go`)
- [ ] Test on different platforms (Linux, Darwin, Windows)