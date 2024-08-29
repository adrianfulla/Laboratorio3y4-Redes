package routing

import (
	"encoding/json"
	"fmt"
)


type Message struct {
    Type    string            `json:"type"`
    From    string            `json:"from"`
    To      string            `json:"to"`
    Hops    int               `json:"hops"`
    Headers []map[string]string `json:"headers"`
    Payload string            `json:"payload"`
}

func (msg *Message) SerializeMessage()(string, error){
	serializedMessage, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to serialize message: %v", err)
	}
	return string(serializedMessage), nil
}

type RoutingAlgorithm interface {
	Name() string
    Initialize(network *Network)            // Initialize the routing algorithm with the network topology
    GetNextHop(source, destination string) (string, error)  // Get the next hop for a message from source to destination
    ProcessIncomingMessage(nodeID string, message *Message) (map[*string]*Message, error)// Handle incoming messages at a node
}

