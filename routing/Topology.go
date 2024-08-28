package routing

import (
	"encoding/json"
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
	Neighbors []string
}

type Network struct {
	Nodes map[string]*Node
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
	}

	for nodeID, neighbors := range topology.Config {
		username, ok := names.Config[nodeID]
		if !ok {
			return nil, fmt.Errorf("node ID %s not found in names file", nodeID)
		}

		network.Nodes[nodeID] = &Node{
			ID:        nodeID,
			Username:  username,
			Neighbors: neighbors,
		}
	}

	return network, nil
}
