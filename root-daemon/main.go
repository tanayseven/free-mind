package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"gopkg.in/zeromq/goczmq.v4"
)

// Message represents the JSON structure for communication
type Message struct {
	Action  string `json:"action"`
	Content string `json:"content"`
}

func FindAvailablePort(startPort, endPort int) (int, error) {
	for port := startPort; port <= endPort; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port, nil
		}
		fmt.Printf("Port %d is not available: %v\n", port, err)
	}
	return 0, fmt.Errorf("no available port found in range %d-%d", startPort, endPort)
}

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

func ProcessMessage(msg Message) {
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

func HostsFilePath() string {
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
	switch runtime.GOOS {
	case "windows":
		return `C:\Windows\System32\drivers\etc\hosts-list-to-be-blocked`
	case "darwin", "linux":
		return "/etc/hosts-list-to-be-blocked"
	default:
		return ""
	}
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

const starter = "########## START OF FREE-MIND BLOCK LIST ##########"
const ender = "########## END OF FREE-MIND BLOCK LIST ##########"

func RestartNetwork() {
	switch runtime.GOOS {
	case "windows":
		exec.Command("ipconfig", "/flushdns")
		break
	case "darwin":
		exec.Command("dscacheutil", "-flushcache")
		break
	case "linux":
		exec.Command("/etc/init.d/networking", "restart")
		exec.Command("/etc/init.d/nscd", "restart")
		exec.Command("/etc/rc.d/nscd", "restart")
		exec.Command("/etc/rc.d/init.d/nscd", "restart")
		break
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

	// Fetch the list of sites to block from the file
	sitesToBlock, err := os.ReadFile(sitesToBlockPath)
	if err != nil {
		log.Printf("Error reading sites to block file: %v", err)
		return
	}

	// Append a starter at the beginning and ending of the list of sites to block
	blockList := fmt.Sprintf("\n%s\n%s\n%s\n", starter, string(sitesToBlock), ender)

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

	// Remove all the lines till the ender marker written into the file
	endIndex := strings.Index(contentStr, ender)
	if endIndex == -1 {
		// End marker not found, just remove from start marker to end of file
		contentStr = contentStr[:startIndex]
	} else {
		// Remove the block including the end marker and the newline after it
		contentStr = contentStr[:startIndex] + contentStr[endIndex+len(ender)+1:]
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
