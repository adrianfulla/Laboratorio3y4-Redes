package main

import (
	"github.com/adrianfulla/Laboratorio3y4-Redes/routing"
	"log"
	"fmt"
)

func CreateNetwork(){

	topologyFile := "topo-example.txt"
	namesFile := "names-example.txt"

	topology, err := routing.ParseTopologyFile(topologyFile)
	if err != nil {
		log.Fatalf("Error parsing topology file: %v", err)
	}

	names, err := routing.ParseNamesFile(namesFile)
	if err != nil {
		log.Fatalf("Error parsing names file: %v", err)
	}

	network, err := routing.BuildNetwork(topology, names)
	if err != nil {
		log.Fatalf("Error building network: %v", err)
	}
	for nodeID, node := range network.Nodes {
		fmt.Printf("Node ID: %s, Username: %s, Neighbors: %v\n", nodeID, node.Username, node.Neighbors)
	}
}