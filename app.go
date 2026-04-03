package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"free-mind/ipc"
)

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
	log.Println("Application started successfully")
}

func FreeMindDaemonSocketPath() string {
	switch runtime.GOOS {
	case "windows":
		return `C:\Windows\System32\drivers\etc\free-mind-daemon-socket-path`
	case "darwin", "linux":
		return "/etc/free-mind-daemon-socket-path"
	default:
		log.Printf("Unsupported platform: %s for socket path", runtime.GOOS)
		return ""
	}
}

func (a *App) ConnectToDaemon() string {
	c := ipc.NewClient()
	err := c.Connect()
	if err != nil {
		log.Printf("Error connecting to daemon: %v", err)
		return fmt.Sprintf("Error: %v", err)
	}
	path := c.Path()
	c.Close()
	return fmt.Sprintf("Connected to daemon at %s", path)
}

func sendMessage(msg *ipc.Message) (*ipc.Message, error) {
	c := ipc.NewClient()
	if err := c.Connect(); err != nil {
		return nil, fmt.Errorf("connect: %v", err)
	}
	defer c.Close()
	if err := c.Send(msg); err != nil {
		return nil, fmt.Errorf("send: %v", err)
	}
	return c.Receive()
}

func (a *App) SendBlockList(list string) bool {
	reply, err := sendMessage(&ipc.Message{Action: "update", Content: list})
	if err != nil {
		log.Printf("Error in SendBlockList: %v", err)
		return false
	}
	return reply.Action == "response" && reply.Content == "Message received successfully"
}

func (a *App) StartBlocking() bool {
	reply, err := sendMessage(&ipc.Message{Action: "start", Content: ""})
	if err != nil {
		log.Printf("Error in StartBlocking: %v", err)
		return false
	}
	return reply.Action == "response" && reply.Content == "Message received successfully"
}

func (a *App) StopBlocking() string {
	reply, err := sendMessage(&ipc.Message{Action: "stop", Content: ""})
	if err != nil {
		return fmt.Sprintf("Error in StopBlocking: %v", err)
	}
	if reply.Action == "response" && reply.Content == "Message received successfully" {
		return "Blocking stopped successfully"
	}
	return "Failed to stop blocking"
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
	log.Println("Starting InstallDaemonBinary function")

	// Extract the daemon binary to a temporary file
	tempBinaryPath, err := a.ExtractDaemonBinary()
	if err != nil {
		return fmt.Sprintf("Error extracting daemon binary: %v", err)
	}
	defer func() {
		// Clean up the temporary file
		if tempBinaryPath != "" {
			os.Remove(tempBinaryPath)
		}
	}()
	log.Printf("Successfully extracted daemon binary to temporary path: %s", tempBinaryPath)

	// Define the destination path
	destPath := a.GetDaemonBinaryDestination()
	if destPath == "" {
		return fmt.Sprintf("Unsupported platform: %s", runtime.GOOS)
	}
	log.Printf("Daemon will be installed to: %s", destPath)

	// Use the appropriate command to install the binary with elevated privileges based on the OS
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		log.Println("Using Windows installation method")
		// On Windows, use PowerShell with elevated privileges
		copyCmd := fmt.Sprintf("Copy-Item -Path '%s' -Destination '%s' -Force", tempBinaryPath, destPath)
		cmd = exec.CommandContext(ctx, "powershell", "-Command", "Start-Process", "powershell",
			"-Verb", "RunAs", "-ArgumentList", fmt.Sprintf("'%s'", copyCmd))
	case "darwin":
		log.Println("Using macOS installation method")
		// On macOS, use osascript to run a shell command with sudo
		copyCmd := fmt.Sprintf("cp %s %s && chmod 755 %s", tempBinaryPath, destPath, destPath)
		appleScript := fmt.Sprintf(`do shell script "%s" with administrator privileges`, copyCmd)
		cmd = exec.CommandContext(ctx, "osascript", "-e", appleScript)
	case "linux":
		log.Println("Using Linux installation method")
		// On Linux, use pkexec
		cmd = exec.CommandContext(ctx, "pkexec", "sh", "-c",
			fmt.Sprintf("cp %s %s && chmod 755 %s", tempBinaryPath, destPath, destPath))
	default:
		return fmt.Sprintf("Unsupported platform: %s", runtime.GOOS)
	}

	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	log.Printf("Executing installation command: %v", cmd.Args)
	err = cmd.Run()
	if err != nil {
		errorMsg := fmt.Sprintf("Error installing daemon binary: %v\nStderr: %s", err, stderr.String())
		log.Println(errorMsg)
		return errorMsg
	}

	log.Printf("Successfully installed daemon to: %s", destPath)
	result := fmt.Sprintf("Daemon binary successfully installed to %s", destPath)
	log.Println(result)
	return result
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
		errorMsg := fmt.Sprintf("Failed to install daemon: %s", installResult)
		log.Println(errorMsg)
		return errorMsg
	}

	// Start the daemon
	var cmd *exec.Cmd
	destPath := a.GetDaemonBinaryDestination()

	log.Printf("Attempting to start daemon from path: %s", destPath)
	
	// Check if the daemon is already running by checking for the socket file
	var socketPath string
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		socketPath = "/tmp/tech.tanay.free-mind.sock"
		log.Printf("DEBUG: Unix socket path set to: %s", socketPath)
	} else {
		// For Windows, use the appropriate named pipe path
		socketPath = `\\.\pipe\tech.tanay.free-mind`
		log.Printf("DEBUG: Windows pipe path set to: %s", socketPath)
	}
	
	_, err := os.Stat(socketPath)
	if err == nil {
		log.Printf("Socket file already exists at %s, checking if it's active", socketPath)
		// Try to connect to see if it's a valid socket
		testClient := ipc.NewClient()
		err = testClient.Connect()
		if err != nil {
			log.Printf("Socket file exists but connection failed: %v. Removing stale socket file.", err)
			os.Remove(socketPath)
		} else {
			log.Printf("Successfully connected to existing socket, daemon is already running")
			testClient.Close()
			return "Daemon is already running"
		}
	} else {
		log.Printf("Socket file check result: %v", err)
	}

	switch runtime.GOOS {
	case "windows":
		log.Println("Using Windows daemon startup method")
		cmd = exec.Command("powershell", "-Command", "Start-Process", destPath, "-WindowStyle", "Hidden")
	case "darwin", "linux":
		log.Println("Using Linux/macOS daemon startup method")
		// Start the daemon with pkexec for elevated privileges
		// Use nohup to ensure the daemon keeps running even if pkexec exits
		cmd = exec.Command("pkexec", "sh", "-c", fmt.Sprintf("nohup %s > /tmp/free-mind-daemon.log 2>&1 &", destPath))
		log.Printf("DEBUG: Starting daemon with command: pkexec sh -c 'nohup %s > /tmp/free-mind-daemon.log 2>&1 &'", destPath)
	default:
		errorMsg := fmt.Sprintf("Unsupported platform: %s", runtime.GOOS)
		log.Println(errorMsg)
		return errorMsg
	}

	// Start the daemon in the background
	log.Printf("Starting daemon command: %v", cmd.Args)
	err = cmd.Start()
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to start daemon: %v", err)
		log.Println(errorMsg)
		return errorMsg
	}

	log.Println("Daemon process started successfully")

	// Wait for the daemon to initialize and create the socket
	log.Println("Waiting for daemon to initialize and create socket...")
	maxRetries := 20
	retryDelay := 500 * time.Millisecond
	socketCreated := false
	
	for i := 0; i < maxRetries; i++ {
		time.Sleep(retryDelay)
		
		// Check if the socket file exists
		if fileInfo, err := os.Stat(socketPath); err == nil {
			// Verify it's a socket file
			if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
				// On Unix systems, check if it's a socket
				if fileInfo.Mode()&os.ModeSocket != 0 {
					log.Printf("Valid socket file detected after %d retries", i+1)
					socketCreated = true
					break
				} else {
					log.Printf("File exists but is not a socket (mode: %v), retrying...", fileInfo.Mode())
				}
			} else {
				// On Windows, we can't easily check if it's a named pipe, so just assume it is
				log.Printf("Socket file detected after %d retries", i+1)
				socketCreated = true
				break
			}
		} else if i == maxRetries-1 {
			log.Printf("Socket file not created after %d retries", maxRetries)
		} else {
			log.Printf("Waiting for socket file (retry %d/%d)...", i+1, maxRetries)
		}
	}
	
	if !socketCreated {
		log.Printf("WARNING: Socket file was not created within the expected time")
	}

	// Check if the daemon is now running
	if a.CheckDaemonInstalled() {
		result := "Daemon successfully installed and started"
		log.Println(result)
		return result
	} else {
		errorMsg := "Daemon was installed but failed to start properly"
		log.Println(errorMsg)
		return errorMsg
	}
}

const hostsStartMarker = "########## START OF FREE-MIND BLOCK LIST ##########"

// CheckBlocking returns true if Free Mind's block list is currently active in /etc/hosts
func (a *App) CheckBlocking() bool {
	hostsPath := a.HostsFilePath()
	if hostsPath == "" {
		return false
	}
	content, err := os.ReadFile(hostsPath)
	if err != nil {
		log.Printf("Error reading hosts file: %v", err)
		return false
	}
	return strings.Contains(string(content), hostsStartMarker)
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
