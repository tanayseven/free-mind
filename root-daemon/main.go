package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

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
			// Port is available, close the listener and return the port
			listener.Close()
			return port, nil
		}
		// If there's an error, it means the port is likely in use or unavailable.
		// We can optionally print the error for debugging.
		// fmt.Printf("Port %d is not available: %v\n", port, err)
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
