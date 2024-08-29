package routing

import (
	"fmt"
)

type Flooding struct {
	network *Network
}

// Initialize initializes the Flooding algorithm with the network topology.
func (f *Flooding) Initialize(network *Network) {
	f.network = network
}

// GetNextHop returns an error because Flooding does not compute specific next hops.
func (f *Flooding) GetNextHop(source, destination string) (string, error) {
	return "", fmt.Errorf("Flooding algorithm does not use specific next hops")
}

// isVisited checks if a nodeID is in the visited list of the message headers
func (f *Flooding) isVisited(nodeID string, headers []map[string]string) bool {
	for _, header := range headers {
		if visited, ok := header["visited"]; ok && visited == nodeID {
			return true
		}
	}
	return false
}

// ProcessIncomingMessage processes incoming messages using the Flooding algorithm.
func (f *Flooding) ProcessIncomingMessage(nodeID string, message *Message) (map[*string]*Message, error) {
	// Check if the message has reached its destination.
	fmt.Println("Processing Message", f.network)
	var recepientNodeID string 
	for id, node := range f.network.Nodes{
		if node.Username == message.To{
			recepientNodeID = id
		}
	}
	messages := make(map[*string]*Message)

	if f.isVisited(nodeID, message.Headers) {
		fmt.Printf("Node %s has already been used", nodeID)
		return nil, nil
	}

	message.Headers = append(message.Headers,map[string]string{"visited":nodeID})

	if recepientNodeID == nodeID {
		fmt.Printf("Message arrived at destination: %s\n", message.Payload)
		messages[nil] = message
	}else{
		// Flood the message to all neighbors except the one it came from.
		neighbors := f.network.Nodes[nodeID].Neighbors
		message.Hops++
		for neighborID := range neighbors {
			// Avoid sending the message back to the node it came from.
			if f.network.Nodes[neighborID].Username == message.From {
				continue
			}else{
				fmt.Printf("Flooding message from %s to %s via %s\n", message.From, message.To, neighborID)
				messages[&neighborID] = message
			}
		}
	}
	

	return messages, nil
}


func (f *Flooding) Name() string{
	return "flooding"
}