package main

import (
	"fmt"
	"log"
	"encoding/json"
	"github.com/adrianfulla/Laboratorio3y4-Redes/routing"
	"github.com/adrianfulla/Proyecto1-Redes/server/xmpp"
	xmppfunctions "github.com/adrianfulla/Proyecto1-Redes/server/xmpp-functions"
	"fyne.io/fyne/v2"
	"path/filepath"
)

func CreateNetwork() routing.Network{

	topologyFiles, err := filepath.Glob("topo-*.txt")
	if err != nil || len(topologyFiles) == 0 {
		log.Fatalf("Error finding topology files: %v", err)
	}

	namesFiles, err := filepath.Glob("names-*.txt")
	if err != nil || len(namesFiles) == 0 {
		log.Fatalf("Error finding names files: %v", err)
	}

	var topology *routing.Topology
	var names *routing.Names

	if len(topologyFiles) > 0 {
		topology, err = routing.ParseTopologyFile(topologyFiles[0])
		if err != nil {
			log.Fatalf("Error parsing topology file: %v", err)
		}
	}

	if len(namesFiles) > 0 {
		names, err = routing.ParseNamesFile(namesFiles[0])
		if err != nil {
			log.Fatalf("Error parsing names file: %v", err)
		}
	}

	network, err := routing.BuildNetwork(topology, names)
	if err != nil {
		log.Fatalf("Error building network: %v", err)
	}
	for nodeID, node := range network.Nodes {
		fmt.Printf("Node ID: %s, Username: %s, Neighbors: %v\n", nodeID, node.Username, node.Neighbors)
	}
	return *network
}


func ParseXMPPMessageBody(xmppMsg *xmpp.Message) (*routing.Message, error) {
    var routingMsg routing.Message

    // Unmarshal the JSON content in the XMPP message body into the routing message struct
    err := json.Unmarshal([]byte(xmppMsg.Body), &routingMsg)
    if err != nil {
        return nil, fmt.Errorf("failed to parse XMPP message body: %v", err)
    }

    return &routingMsg, nil
}

func ProcessMessage(handler *xmpp.XMPPHandler, nodeID string, msg *routing.Message, routingTable routing.Network, fullMsg *xmpp.Message){
	// Use the selected routing algorithm to handle the message
	messages, err := routingTable.Algorithm.ProcessIncomingMessage(nodeID, msg)
	if err != nil{
		fmt.Println("Error processing message: ",err)
	}else {
		for nextHop, updatedMessage := range messages{
			if nextHop != nil {
				nextHopUser := routingTable.Nodes[*nextHop].Username
				serializedMessage, err := updatedMessage.SerializeMessage()
				if err != nil{
					fmt.Println("Error serializing message: ",err)  
				}
				err = xmppfunctions.SendMessage(handler, nextHopUser, serializedMessage)
				if err != nil{
					fmt.Println("Error sending message: ", err)
				}
			}else {
				// Handle the message if it is intended for this node (e.g., display it in the chat window)
				if chatWindow, ok := handler.ChatWindows[msg.From]; ok && chatWindow != nil {
					chatWindow.AddMessage(fullMsg)
					fyne.CurrentApp().SendNotification(&fyne.Notification{
						Title:   "New Message",
						Content: fmt.Sprintf("%s: %s", msg.From, fullMsg.Body),
					})
				} else {
					if len(fullMsg.Body) > 0 {
						log.Printf("No chat window open for %s, queueing message", msg.From)
						handler.MessageQueue[msg.From] = append(handler.MessageQueue[msg.From], fullMsg)
		
						// Send notification for each queued message
						fyne.CurrentApp().SendNotification(&fyne.Notification{
							Title:   "New Message",
							Content: fmt.Sprintf("%s: %s", msg.From, fullMsg.Body),
						})
					}
				}
			}
		}
	}
	
	
}