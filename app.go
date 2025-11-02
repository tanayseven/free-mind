package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gopkg.in/zeromq/goczmq.v4"
)

// Message represents the JSON structure for communication
type Message struct {
	Action  string `json:"action"`
	Content string `json:"content"`
}

// Global ZMQ dealer
var dealer *goczmq.Sock

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func FreeMindDaemonPortPath() string {
	switch runtime.GOOS {
	case "windows":
		return `C:\Windows\System32\drivers\etc\free-mind-daemon-port-number`
	case "darwin", "linux":
		return "/etc/free-mind-daemon-port-number"
	default:
		return ""
	}
}

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

func (a *App) HostsFilePath() string {
	switch runtime.GOOS {
	case "windows":
		return `C:\Windows\System32\drivers\etc\hosts`
	case "darwin", "linux":
		return "/etc/hosts"
	default:
		return ""
	}
}

func (a *App) WriteToHostFile() string {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pkexec", "sh", "-c", "echo \"# test\" >> /etc/hosts")
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	return out.String() + "\n" + stderr.String()
}
