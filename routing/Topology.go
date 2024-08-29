package routing

import (
	"encoding/json"
	"container/heap"
	"math"
	"fmt"
	"io/ioutil"
	// "log"
	"os"
)

// Topology represents the structure of the topology file.
type Topology struct {
	Type   string               `json:"type"`
	Config map[string][]string   `json:"config"`
}

// Names represents the structure of the names file.
type Names struct {
	Type   string            `json:"type"`
	Config map[string]string `json:"config"`
}

// Node represents each node in the network.
type Node struct {
	ID        string
	Username  string
	Neighbors map[string]int
}

type Network struct {
	Nodes map[string]*Node
	Algorithm RoutingAlgorithm
}

func ReadJSONFile(filename string, v interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, v)
}

func ParseTopologyFile(filename string) (*Topology, error) {
	var topology Topology
	err := ReadJSONFile(filename, &topology)
	if err != nil {
		return nil, fmt.Errorf("failed to parse topology file: %w", err)
	}
	return &topology, nil
}

func ParseNamesFile(filename string) (*Names, error) {
	var names Names
	err := ReadJSONFile(filename, &names)
	if err != nil {
		return nil, fmt.Errorf("failed to parse names file: %w", err)
	}
	return &names, nil
}

// BuildNetwork builds a Network struct from the given topology and names.
func BuildNetwork(topology *Topology, names *Names) (*Network, error) {
	network := &Network{
		Nodes: make(map[string]*Node),
		Algorithm: &Flooding{},
	}

	for nodeID, neighbors := range topology.Config {
		username, ok := names.Config[nodeID]
		if !ok {
			return nil, fmt.Errorf("node ID %s not found in names file", nodeID)
		}

		neighbors_with_weights := make(map[string]int)

		for _,neighbor := range neighbors{
			neighbors_with_weights[neighbor] = 1
		}

		network.Nodes[nodeID] = &Node{
			ID:        nodeID,
			Username:  username,
			Neighbors: neighbors_with_weights,
		}
	}

	network.Algorithm.Initialize(network)

	return network, nil
}

// Dijkstra's algorithm implementation to find shortest paths.
func (net *Network) Dijkstra(sourceID string) (map[string]string, map[string]int) {
	dist := make(map[string]int)
	prev := make(map[string]string)
	unvisited := &MinHeap{}

	heap.Init(unvisited)

	for nodeID := range net.Nodes {
		dist[nodeID] = math.MaxInt64
		heap.Push(unvisited, &Item{Value: nodeID, Priority: dist[nodeID]})
	}

	dist[sourceID] = 0
	heap.Push(unvisited, &Item{Value: sourceID, Priority: dist[sourceID]})

	for unvisited.Len() > 0 {
		u := heap.Pop(unvisited).(*Item).Value

		for neighborID, cost := range net.Nodes[u].Neighbors {
			alt := dist[u] + cost
			if alt < dist[neighborID] {
				dist[neighborID] = alt
				prev[neighborID] = u
				heap.Push(unvisited, &Item{Value: neighborID, Priority: dist[neighborID]})
			}
		}
	}

	// Create a next-hop map for the source node
	nextHop := make(map[string]string)
	for nodeID := range net.Nodes {
		if nodeID != sourceID {
			at := nodeID
			for prev[at] != sourceID && prev[at] != "" {
				at = prev[at]
			}
			if prev[at] == sourceID {
				nextHop[nodeID] = at
			} else {
				nextHop[nodeID] = nodeID
			}
		}
	}

	return nextHop, dist
}


// MinHeap to implement priority queue for Dijkstra's algorithm
type Item struct {
    Value    string
    Priority int
    Index    int
}

type MinHeap []*Item

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].Priority < h[j].Priority }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].Index = i; h[j].Index = j }

func (h *MinHeap) Push(x interface{}) {
    item := x.(*Item)
    item.Index = len(*h)
    *h = append(*h, item)
}

func (h *MinHeap) Pop() interface{} {
    old := *h
    n := len(old)
    item := old[n-1]
    old[n-1] = nil // avoid memory leak
    item.Index = -1
    *h = old[0 : n-1]
    return item
}
