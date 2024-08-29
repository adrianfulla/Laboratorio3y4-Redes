package routing

import (
	"fmt"
)

type LinkStateRouting struct {
    network *Network
    shortestPaths map[string]map[string]string // [source][destination] -> nextHop
}

// computeShortestPaths computes the shortest path from the source to all other nodes using Dijkstra's algorithm.
func (lsr *LinkStateRouting) computeShortestPaths(source string) map[string]string {
	nextHop, _ := lsr.network.Dijkstra(source)
	return nextHop
}

// Initialize initializes the Link State Routing algorithm.
func (lsr *LinkStateRouting) Initialize(network *Network) {
	lsr.network = network
	lsr.shortestPaths = make(map[string]map[string]string)

	// Compute shortest paths for each node in the network.
	for nodeID := range network.Nodes {
		lsr.shortestPaths[nodeID] = lsr.computeShortestPaths(nodeID)
	}
}

// GetNextHop returns the next hop for a message from the source to the destination.
func (lsr *LinkStateRouting) GetNextHop(source, destination string) (string, error) {
	fmt.Printf("Get next hop from %s to %s\n",source,destination )
	if nextHop, ok := lsr.shortestPaths[source][destination]; ok {
		return nextHop, nil
	}
	return "", fmt.Errorf("no path from %s to %s", source, destination)
}

// ProcessIncomingMessage processes incoming messages, forwarding them to the next hop or handling them if they are at their destination.
func (lsr *LinkStateRouting) ProcessIncomingMessage(nodeID string, message *Message) (map[*string]*Message, error){
	
	var recepientNodeID string 
	for id, node := range lsr.network.Nodes{
		if node.Username == message.To{
			recepientNodeID = id
		}
	}
	messages := make(map[*string]*Message)
	if recepientNodeID == nodeID {
		fmt.Printf("Message arrived at destination: %s\n", message.Payload)
		messages[nil] = message
	} else {
		nextHop, err := lsr.GetNextHop(nodeID, recepientNodeID)
		if err != nil {
			fmt.Printf("No path found for message: %v\n", err)
			return nil, err
		} else {
			message.Hops++
			fmt.Printf("Forwarding message from %s to %s via %s\n", message.From, message.To, nextHop)
			messages[&nextHop] = message 
		}
	}
	return messages, nil
}