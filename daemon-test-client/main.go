package main

import (
	"encoding/json"
	"fmt"
	"log"

	"gopkg.in/zeromq/goczmq.v4"
)

// Message represents the JSON structure for communication
type Message struct {
	Action  string `json:"action"`
	Content string `json:"content"`
}

func main() {
	dealer, err := goczmq.NewDealer(fmt.Sprintf("tcp://127.0.0.1:%d", 5500))
	if err != nil {
		log.Fatal(err)
	}
	defer dealer.Destroy()
	// Create a message to send
	msg := Message{
		Action:  "start",
		Content: "foobar",
	}

	// Marshal the message to JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}

	// Send the JSON data
	err = dealer.SendFrame(jsonData, goczmq.FlagNone)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("dealer sent message with action: '%s', content: '%s'", msg.Action, msg.Content)

	// Receive the reply
	reply, err := dealer.RecvMessage()
	if err != nil {
		log.Fatal(err)
	}

	// Parse the received JSON response
	var responseMsg Message
	err = json.Unmarshal(reply[0], &responseMsg)
	if err != nil {
		log.Printf("Error unmarshaling response JSON: %v", err)
		log.Printf("Raw response: '%s'", string(reply[0]))
	} else {
		log.Printf("dealer received response - action: '%s', content: '%s'",
			responseMsg.Action, responseMsg.Content)
	}
}
