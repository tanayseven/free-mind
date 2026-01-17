package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// GetDaemonBinaryPath returns the path to the daemon binary for the current platform
func (a *App) GetDaemonBinaryPath() string {
	switch runtime.GOOS {
	case "windows":
		return "build/bin/free-mind-daemon-windows.exe"
	case "darwin":
		return "build/bin/free-mind-daemon-darwin"
	case "linux":
		return "build/bin/free-mind-daemon-linux"
	default:
		return ""
	}
}

// GetDaemonBinaryDestination returns the destination path for the daemon binary
func (a *App) GetDaemonBinaryDestination() string {
	switch runtime.GOOS {
	case "windows":
		return `C:\Windows\System32\free-mind-daemon.exe`
	case "darwin", "linux":
		return "/usr/bin/free-mind-daemon"
	default:
		return ""
	}
}

// ExtractDaemonBinary extracts the embedded daemon binary to a temporary file
func (a *App) ExtractDaemonBinary() (string, error) {
	// Get the appropriate binary path for the current platform
	binaryPath := a.GetDaemonBinaryPath()
	if binaryPath == "" {
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// Read the binary from the embedded filesystem
	binaryData, err := daemonBinaries.ReadFile(binaryPath)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded binary: %v", err)
	}

	// Create a temporary file to store the extracted binary
	tempFile, err := ioutil.TempFile("", "free-mind-daemon-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer tempFile.Close()

	// Write the binary data to the temporary file
	if _, err := tempFile.Write(binaryData); err != nil {
		return "", fmt.Errorf("failed to write binary data to temporary file: %v", err)
	}

	// Make the temporary file executable
	if err := os.Chmod(tempFile.Name(), 0755); err != nil {
		return "", fmt.Errorf("failed to make temporary file executable: %v", err)
	}

	return tempFile.Name(), nil
}

// InstallDaemonBinary installs the daemon binary to the appropriate system location with elevated privileges
func (a *App) InstallDaemonBinary() string {
	// Extract the daemon binary to a temporary file
	tempBinaryPath, err := a.ExtractDaemonBinary()
	if err != nil {
		return fmt.Sprintf("Error extracting daemon binary: %v", err)
	}
	defer os.Remove(tempBinaryPath) // Clean up the temporary file

	// Define the destination path
	destPath := a.GetDaemonBinaryDestination()
	if destPath == "" {
		return fmt.Sprintf("Unsupported platform: %s", runtime.GOOS)
	}

	// Use the appropriate command to install the binary with elevated privileges based on the OS
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// On Windows, use PowerShell with elevated privileges
		copyCmd := fmt.Sprintf("Copy-Item -Path '%s' -Destination '%s' -Force", tempBinaryPath, destPath)
		cmd = exec.CommandContext(ctx, "powershell", "-Command", "Start-Process", "powershell",
			"-Verb", "RunAs", "-ArgumentList", fmt.Sprintf("'%s'", copyCmd))
	case "darwin":
		// On macOS, use osascript to run a shell command with sudo
		copyCmd := fmt.Sprintf("cp %s %s && chmod 755 %s", tempBinaryPath, destPath, destPath)
		appleScript := fmt.Sprintf(`do shell script "%s" with administrator privileges`, copyCmd)
		cmd = exec.CommandContext(ctx, "osascript", "-e", appleScript)
	case "linux":
		// On Linux, use pkexec
		cmd = exec.CommandContext(ctx, "pkexec", "sh", "-c",
			fmt.Sprintf("cp %s %s && chmod 755 %s", tempBinaryPath, destPath, destPath))
	default:
		return fmt.Sprintf("Unsupported platform: %s", runtime.GOOS)
	}

	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Sprintf("Error installing daemon binary: %v\n%s", err, stderr.String())
	}

	return fmt.Sprintf("Daemon binary successfully installed to %s", destPath)
}

// InstallDaemonWithOneClick is a frontend-exposed method to install the daemon with one click
func (a *App) InstallDaemonWithOneClick() string {
	log.Println("Installing daemon binary with one click...")

	// Check if we're running with sufficient privileges
	result := a.InstallDaemonBinary()

	// Log the result
	log.Println("Installation result:", result)

	return result
}

// CheckDaemonInstalled checks if the daemon is installed and running
func (a *App) CheckDaemonInstalled() bool {
	log.Println("Checking if daemon is installed and running...")

	// Check if the daemon binary exists
	destPath := a.GetDaemonBinaryDestination()
	if destPath == "" {
		log.Printf("Unsupported platform: %s", runtime.GOOS)
		return false
	}

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

	// Check if the daemon port file exists
	portPath := FreeMindDaemonPortPath()
	_, err = os.Stat(portPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Daemon port file not found at:", portPath)
		} else {
			log.Printf("Error checking daemon port file: %v", err)
		}
		return false
	}

	// Try to connect to the daemon
	portData, err := os.ReadFile(portPath)
	if err != nil {
		log.Printf("Error reading daemon port file: %v", err)
		return false
	}

	// Convert port string to integer
	portStr := strings.TrimSpace(string(portData))
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

// InstallAndStartDaemon installs the daemon if not already installed and starts it
func (a *App) InstallAndStartDaemon() string {
	log.Println("Installing and starting daemon...")

	// Check if daemon is already installed
	if a.CheckDaemonInstalled() {
		log.Println("Daemon is already installed and running")
		return "Daemon is already installed and running"
	}

	// Install the daemon
	installResult := a.InstallDaemonBinary()
	log.Println("Installation result:", installResult)

	// Check if installation was successful
	if !strings.Contains(installResult, "successfully") {
		return fmt.Sprintf("Failed to install daemon: %s", installResult)
	}

	// Start the daemon
	var cmd *exec.Cmd
	destPath := a.GetDaemonBinaryDestination()

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("powershell", "-Command", "Start-Process", destPath, "-WindowStyle", "Hidden")
	case "darwin", "linux":
		// Start the daemon with pkexec for elevated privileges
		cmd = exec.Command("pkexec", destPath)
	default:
		return fmt.Sprintf("Unsupported platform: %s", runtime.GOOS)
	}

	// Start the daemon in the background
	err := cmd.Start()
	if err != nil {
		return fmt.Sprintf("Failed to start daemon: %v", err)
	}

	// Wait a moment for the daemon to initialize
	time.Sleep(2 * time.Second)

	// Check if the daemon is now running
	if a.CheckDaemonInstalled() {
		return "Daemon successfully installed and started"
	} else {
		return "Daemon was installed but failed to start properly"
	}
}

// CheckAndInstallDaemon checks if the daemon is installed and running, and installs it if not
func (a *App) CheckAndInstallDaemon() string {
	log.Println("Checking and installing daemon if needed...")

	// Check if daemon is already installed and running
	if a.CheckDaemonInstalled() {
		log.Println("Daemon is already installed and running")
		return "Daemon is already installed and running"
	}

	// Install and start the daemon
	result := a.InstallAndStartDaemon()
	log.Println("Installation and startup result:", result)

	return result
}
