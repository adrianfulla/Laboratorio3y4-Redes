package routing

import (
	"fmt"
)

type DistanceVectorRouting struct {
	network       *Network
	distanceTable map[string]map[string]int    // [node][destination] -> cost
	nextHopTable  map[string]map[string]string // [node][destination] -> nextHop
}

// Initialize initializes the Distance Vector Routing algorithm.
func (dvr *DistanceVectorRouting) Initialize(network *Network) {
	dvr.network = network
	dvr.distanceTable = make(map[string]map[string]int)
	dvr.nextHopTable = make(map[string]map[string]string)

	// Initialize distance tables with direct neighbor information.
	for nodeID, node := range network.Nodes {
		dvr.distanceTable[nodeID] = make(map[string]int)
		dvr.nextHopTable[nodeID] = make(map[string]string)

		for neighborID, cost := range node.Neighbors {
			dvr.distanceTable[nodeID][neighborID] = cost
			dvr.nextHopTable[nodeID][neighborID] = neighborID
		}
		dvr.distanceTable[nodeID][nodeID] = 0 // Distance to self is 0
	}

	// Run the Bellman-Ford algorithm to initialize the tables
	dvr.BellmanFord()
}

// BellmanFord runs the Bellman-Ford algorithm to compute the shortest paths.
func (dvr *DistanceVectorRouting) BellmanFord() {
	// Iterate until no changes are made (convergence)
	updated := true

	for updated {
		updated = false

		for nodeID := range dvr.network.Nodes {
			for neighborID := range dvr.network.Nodes[nodeID].Neighbors {
				for destinationID := range dvr.network.Nodes {
					if destinationID == nodeID {
						continue // Skip if it's the same node
					}

					newCost := dvr.distanceTable[neighborID][destinationID] + dvr.network.Nodes[nodeID].Neighbors[neighborID]
					if currentCost, exists := dvr.distanceTable[nodeID][destinationID]; !exists || newCost < currentCost {
						dvr.distanceTable[nodeID][destinationID] = newCost
						dvr.nextHopTable[nodeID][destinationID] = neighborID
						updated = true
					}
				}
			}
		}
	}
}

// GetNextHop returns the next hop for a message from the source to the destination.
func (dvr *DistanceVectorRouting) GetNextHop(source, destination string) (string, error) {
	if nextHop, ok := dvr.nextHopTable[source][destination]; ok {
		return nextHop, nil
	}
	return "", fmt.Errorf("no path from %s to %s", source, destination)
}

// ProcessIncomingMessage processes incoming messages, forwarding them to the next hop or handling them if they are at their destination.
func (dvr *DistanceVectorRouting) ProcessIncomingMessage(nodeID string, message *Message) (map[*string]*Message, error) {
	var recipientNodeID string
	for id, node := range dvr.network.Nodes {
		if node.Username == message.To {
			recipientNodeID = id
		}
	}
	messages := make(map[*string]*Message)
	if recipientNodeID == nodeID {
		fmt.Printf("Message arrived at destination: %s\n", message.Payload)
		messages[nil] = message
	} else {
		nextHop, err := dvr.GetNextHop(nodeID, recipientNodeID)
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
