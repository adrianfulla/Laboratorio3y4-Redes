package main

import (
	"fmt"
	"log"
	"encoding/json"
	"github.com/adrianfulla/Laboratorio3y4-Redes/routing"
	"github.com/adrianfulla/Proyecto1-Redes/server/xmpp"
	xmppfunctions "github.com/adrianfulla/Proyecto1-Redes/server/xmpp-functions"
	"fyne.io/fyne/v2"
)

func CreateNetwork() routing.Network{

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
	return *network
}

func SendUsingFlooding(handler *xmpp.XMPPHandler, recepient string, message string, routingTable routing.Network) error{
	return nil
}
func SendUsingLinkState(handler *xmpp.XMPPHandler, recepient string, message string, routingTable routing.Network) error{
	return nil
}
func SendUsingDistanceVector(handler *xmpp.XMPPHandler, recepient string, message string, routingTable routing.Network) error{
	return nil
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