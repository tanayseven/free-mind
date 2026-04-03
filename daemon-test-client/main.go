package main

import (
	"log"

	"free-mind/ipc"
)

func main() {
	// Create a new IPC client
	client := ipc.NewClient()
	
	// Connect to the daemon
	err := client.Connect()
	if err != nil {
		log.Fatalf("Error connecting to daemon: %v", err)
	}
	defer client.Close()
	
	// Create a message to send
	msg := &ipc.Message{
		Action:  "start",
		Content: "foobar",
	}
	
	// Send the message
	err = client.Send(msg)
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}
	log.Printf("Sent message with action: '%s', content: '%s'", msg.Action, msg.Content)
	
	// Receive the reply
	reply, err := client.Receive()
	if err != nil {
		log.Fatalf("Error receiving response: %v", err)
	}
	
	log.Printf("Received response - action: '%s', content: '%s'", reply.Action, reply.Content)
}
