package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"free-mind/ipc"
)

func main() {
	log.Printf("DEBUG: Starting root daemon process")

	// Create a new IPC server
	server := ipc.NewServer()
	if server == nil {
		log.Fatalf("ERROR: Failed to create IPC server")
		return
	}
	log.Printf("DEBUG: IPC server created successfully")

	// Write the socket/pipe path to the correct file (optional, as the path is fixed)
	socketPath := server.Path()
	log.Printf("DEBUG: Server socket path: %s", socketPath)

	socketPathFile := FreeMindDaemonSocketPath()
	log.Printf("DEBUG: Socket path file location: %s", socketPathFile)

	err := os.WriteFile(socketPathFile, []byte(socketPath), 0644)
	if err != nil {
		log.Printf("Error writing socket path to file: %v", err)
		// Continue execution even if writing path fails
	} else {
		log.Printf("Daemon socket path %s written to %s", socketPath, socketPathFile)
	}

	// Check if the socket directory exists and is writable
	socketDir := "/tmp"
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		// Ensure the socket file doesn't already exist
		if _, err := os.Stat(socketPath); err == nil {
			log.Printf("Socket file %s already exists, removing it", socketPath)
			if err := os.Remove(socketPath); err != nil {
				log.Printf("Failed to remove existing socket file: %v", err)
			} else {
				log.Printf("Successfully removed existing socket file")
			}
		}

		dirInfo, err := os.Stat(socketDir)
		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("Socket directory %s does not exist, attempting to create it", socketDir)
				err = os.MkdirAll(socketDir, 0755)
				if err != nil {
					log.Printf("Failed to create socket directory %s: %v", socketDir, err)
				} else {
					log.Printf("Successfully created socket directory %s", socketDir)
				}
			} else {
				log.Printf("Error checking socket directory %s: %v", socketDir, err)
			}
		} else if !dirInfo.IsDir() {
			log.Printf("Socket path %s exists but is not a directory", socketDir)
		} else {
			log.Printf("Socket directory %s exists", socketDir)

			// Check if we have write permission to the directory
			testFile := socketDir + "/test-write-permission"
			err = os.WriteFile(testFile, []byte("test"), 0644)
			if err != nil {
				log.Printf("No write permission to socket directory %s: %v", socketDir, err)
			} else {
				os.Remove(testFile)
				log.Printf("Socket directory %s exists and is writable", socketDir)
			}
		}
	}

	// Start listening for connections
	log.Printf("DEBUG: Starting to listen on socket %s", socketPath)
	err = server.Listen()
	if err != nil {
		log.Fatalf("Error starting IPC server: %v", err)
		return
	}
	defer server.Close()

	// Verify the socket file exists after Listen()
	if _, statErr := os.Stat(socketPath); statErr != nil {
		log.Printf("DEBUG: After Listen() - Socket file check error: %v", statErr)
	} else {
		log.Printf("DEBUG: After Listen() - Socket file exists at %s", socketPath)
	}

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

func ProcessMessage(msg *ipc.Message) {
	switch msg.Action {
	case "start":
		StartBlocking()
		break
	case "update":
		UpdateSitesToBeBlocked(msg.Content)
		break
	case "stop":
		StopBlocking()
		break
	}
}

// hostsFilePathOverride and sitesToBlockPathOverride allow tests to redirect
// file operations to a temporary directory instead of /etc.
var hostsFilePathOverride string
var sitesToBlockPathOverride string

func HostsFilePath() string {
	if hostsFilePathOverride != "" {
		return hostsFilePathOverride
	}
	switch runtime.GOOS {
	case "windows":
		return `C:\Windows\System32\drivers\etc\hosts`
	case "darwin", "linux":
		return "/etc/hosts"
	default:
		return ""
	}
}

func SitesToBlockPath() string {
	if sitesToBlockPathOverride != "" {
		return sitesToBlockPathOverride
	}
	switch runtime.GOOS {
	case "windows":
		return `C:\Windows\System32\drivers\etc\hosts-list-to-be-blocked`
	case "darwin", "linux":
		return "/etc/hosts-list-to-be-blocked"
	default:
		return ""
	}
}

const starter = "########## START OF FREE-MIND BLOCK LIST ##########"
const ender = "########## END OF FREE-MIND BLOCK LIST ##########"

func RestartNetwork() {
	switch runtime.GOOS {
	case "windows":
		exec.Command("ipconfig", "/flushdns").Run()
	case "darwin":
		exec.Command("dscacheutil", "-flushcache").Run()
		exec.Command("killall", "-HUP", "mDNSResponder").Run()
	case "linux":
		exec.Command("/etc/init.d/nscd", "restart").Run()
		exec.Command("/etc/rc.d/nscd", "restart").Run()
		exec.Command("/etc/rc.d/init.d/nscd", "restart").Run()
	}
}

func StartBlocking() {
	// Fetch the host path
	hostsPath := HostsFilePath()
	sitesToBlockPath := SitesToBlockPath()

	// Read and create a backup of the hosts file
	hostsContent, err := os.ReadFile(hostsPath)
	if err != nil {
		log.Printf("Error reading hosts file: %v", err)
		return
	}

	// Don't add a second block if one is already present
	if strings.Contains(string(hostsContent), starter) {
		log.Printf("Block list already present in hosts file, skipping")
		return
	}

	// Fetch the list of sites to block from the file
	sitesToBlock, err := os.ReadFile(sitesToBlockPath)
	if err != nil {
		log.Printf("Error reading sites to block file: %v", err)
		return
	}

	// Append a starter at the beginning and ending of the list of sites to block
	blockList := fmt.Sprintf("\n%s\n%s\n%s\n\n", starter, strings.TrimRight(string(sitesToBlock), "\n"), ender)

	// Read the hosts file and append the above assembled list to the hosts
	newHostsContent := string(hostsContent) + blockList

	// Write the hosts file back to the actual file
	err = os.WriteFile(hostsPath, []byte(newHostsContent), 0644)
	if err != nil {
		log.Printf("Error writing hosts file: %v", err)
		return
	}

	// Restart network commands
	RestartNetwork()

	log.Println("Site blocking started successfully")
}

func UpdateSitesToBeBlocked(content string) {
	// Write the files to write to hosts file
	sitesToBlockPath := SitesToBlockPath()

	// Split the list of sites to block into multiple lines
	sites := strings.Split(content, ",")
	var formattedSites strings.Builder

	for _, site := range sites {
		site = strings.TrimSpace(site)
		if site != "" {
			formattedSites.WriteString("127.0.0.1 " + site + "\n")
		}
	}

	err := os.WriteFile(sitesToBlockPath, []byte(formattedSites.String()), 0644)
	if err != nil {
		log.Printf("Error writing sites to block file: %v", err)
		return
	}

	log.Println("Sites to block updated successfully")
}

func StopBlocking() {
	// Fetch the host path
	hostsPath := HostsFilePath()

	// Fetch the hosts file contents
	hostsContent, err := os.ReadFile(hostsPath)
	if err != nil {
		log.Printf("Error reading hosts file: %v", err)
		return
	}

	// Match the starter marker in the contents
	contentStr := string(hostsContent)
	startIndex := strings.Index(contentStr, starter)

	if startIndex == -1 {
		// No block list found, nothing to do
		log.Println("No block list found in hosts file")
		return
	}

	// StartBlocking always prepends a '\n' separator before the marker.
	// Include it in the removal so the file is byte-for-byte identical afterwards.
	blockStart := startIndex
	if blockStart > 0 && contentStr[blockStart-1] == '\n' {
		blockStart -= 2
	}

	// Remove all the lines till the ender marker written into the file
	endIndex := strings.Index(contentStr, ender)
	if endIndex == -1 {
		// End marker not found, just remove from block start to end of file
		contentStr = contentStr[:blockStart]
	} else {
		// Remove the block including the end marker and the newline after it
		contentStr = contentStr[:blockStart] + contentStr[endIndex+len(ender)+1:]
	}

	// Write the hosts file back to the actual file
	err = os.WriteFile(hostsPath, []byte(contentStr), 0644)
	if err != nil {
		log.Printf("Error writing hosts file: %v", err)
		return
	}

	// Restart network commands
	RestartNetwork()

	log.Println("Site blocking stopped successfully")
}
