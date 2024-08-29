package main

import (
	"fmt"
	"log"
	"strings"
	// "time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	// "github.com/adrianfulla/Laboratorio3y4-Redes/routing"
	"github.com/adrianfulla/Laboratorio3y4-Redes/routing"
	xmpp "github.com/adrianfulla/Proyecto1-Redes/server/xmpp"
	xmppfunctions "github.com/adrianfulla/Proyecto1-Redes/server/xmpp-functions"
)

func ShowLoginWindow() {
    routingTable := CreateNetwork()

    myApp := app.New()
    myWindow := myApp.NewWindow("XMPP Chat Client")

    serverEntry := widget.NewEntry()
    serverEntry.SetPlaceHolder("Server (e.g., alumchat.lol:5222)")

    usernameEntry := widget.NewEntry()
    usernameEntry.SetPlaceHolder("Username")

    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Password")

    loginButton := widget.NewButton("Login", func() {
        server := serverEntry.Text
        username := usernameEntry.Text
        password := passwordEntry.Text

        hostPort := strings.Split(server, ":")
        if len(hostPort) != 2 {
            dialog.ShowError(fmt.Errorf("invalid server format"), myWindow)
            return
        }

        handler, err := xmppfunctions.Login(hostPort[0], hostPort[1], username, password)
        if err != nil {
            log.Printf("Login failed: %v", err)
            dialog.ShowError(err, myWindow)
            return
        }

        ShowNodesWindow(myApp, handler, routingTable)
        myWindow.Close()
    })

    createAccountButton := widget.NewButton("Create Account", func() {
        ShowCreateAccountDialog(myApp, myWindow)
    })

    myWindow.SetContent(container.NewVBox(
        widget.NewLabel("Login to XMPP Server"),
        serverEntry,
        usernameEntry,
        passwordEntry,
        loginButton,
        createAccountButton,
    ))

    myWindow.ShowAndRun()
}


// ShowContactsWindow displays the user's contact list.
func ShowChatWindow(app fyne.App, handler *xmpp.XMPPHandler, recipient string, contact xmppfunctions.Contact, routingTable routing.Network, nodeID string) *xmpp.ChatWindow {
    chatWindow := app.NewWindow("Chat with " + recipient + " Node: " +contact.Status)

    messageEntry := widget.NewEntry()
    messageEntry.SetPlaceHolder("Type your message...")

    chatContent := container.NewVBox()

    if queuedMessages, ok := handler.MessageQueue[recipient]; ok {
        for _, msg := range queuedMessages {
            parsedMessage, err := ParseXMPPMessageBody(msg)
            if err != nil{
                fmt.Println("Error parsing message: ", err)
            }else{
                serializedMessage, err := parsedMessage.SerializeMessage()
                if err != nil {
                    fmt.Println("Error serializing message: ", err)
                }else{
                    chatContent.Add(widget.NewLabel(fmt.Sprintf("%s: %s", strings.Split(parsedMessage.From, "/")[0],serializedMessage )))
                } 
            }
        }
        delete(handler.MessageQueue, recipient) 
    }

    sendMessageButton := widget.NewButton("Send", func() {
        message := messageEntry.Text
        if message != "" {
            msg := &routing.Message{
                Type:    "message",
                From:    handler.Username + "@"+ strings.Split(handler.Server,":")[0],
                To:      recipient,
                Hops:    0,
                Headers: []map[string]string{},
                Payload: message,
            }
            msg.Headers = append(msg.Headers,map[string]string{"algorithm":routingTable.Algorithm.Name()})
            // Use the selected routing algorithm to handle the message
            messages, err := routingTable.Algorithm.ProcessIncomingMessage(nodeID, msg)
            if err != nil{
                fmt.Println("Error processing message: ",err)
            }else{
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
                        serializedMessage, err := updatedMessage.SerializeMessage()
                        if err != nil{
                            fmt.Println("Error serializing message: ",err)  
                        }
                        chatContent.Add(widget.NewLabel(updatedMessage.From +": "+ serializedMessage))
                    }
                }
            }
             
            chatContent.Add(widget.NewLabel("Me: " + message))
            messageEntry.SetText("")
        }
    })

    // Button to show contact details
    contactDetailsButton := widget.NewButton("Contact Details", func() {
        ShowContactDetailsWindow(app, handler, contact)
    })

    // Layout: message entry, send button, and contact details button
    messageRow := container.New(layout.NewGridLayoutWithColumns(2), messageEntry, sendMessageButton)
    buttonRow := container.NewHBox(contactDetailsButton)

    chatWindow.SetContent(container.NewBorder(
        chatContent,
        container.NewVBox(messageRow, buttonRow),
        nil, nil,
    ))

    // Handle incoming messages in a separate goroutine
    go func() {
        for {
            err := handler.HandleIncomingStanzas()
            if err != nil {
                log.Printf("Error handling stanzas: %v", err)
                continue
            }

            chatContent.Add(widget.NewLabel("Received a new message...")) 
            chatWindow.Content().Refresh()
        }
    }()

    chatWindow.Resize(fyne.NewSize(400, 500))
    chatWindow.Show()

    chatWindow.SetOnClosed(func() {
        delete(handler.ChatWindows, recipient)
    })

    return &xmpp.ChatWindow{
        Window:      chatWindow,
        ChatContent: chatContent,
        Handler:     handler,
        Recipient:   recipient,
    }
}

func ShowUserSettingsWindow(app fyne.App, handler *xmpp.XMPPHandler, routingTable *routing.Network) {
    settingsWindow := app.NewWindow("User Settings")

    // Create a variable to store the selected routing algorithm
    selectedAlgorithmName := "Flooding"

    routingOptions := widget.NewRadioGroup([]string{"Flooding", "Link State Routing", "Distance Vector Routing"}, func(value string) {
        selectedAlgorithmName = value
        log.Printf("Selected routing algorithm: %s", value)
        // Initialize the chosen routing algorithm
        switch selectedAlgorithmName {
        case "Flooding":
            routingTable.Algorithm = &routing.Flooding{}
        case "Link State Routing":
            routingTable.Algorithm = &routing.LinkStateRouting{}
        case "Distance Vector Routing":
            routingTable.Algorithm = &routing.DistanceVectorRouting{}
        default:
            routingTable.Algorithm = &routing.Flooding{}
            // selectedAlgorithm = nil
        }
        if routingTable.Algorithm != nil{
            routingTable.Algorithm.Initialize(routingTable)
        }
        
    })
    routingOptions.SetSelected(selectedAlgorithmName)

    // Buttons for Logout and Delete Account
    logoutButton := widget.NewButton("Logout", func() {
        err := xmppfunctions.Logout(handler)
        if err != nil {
            log.Printf("Logout failed: %v", err)
            dialog.ShowError(err, settingsWindow)
        } else {
            log.Println("Logged out successfully")
            CloseAllWindows(app)
            settingsWindow.Close()
            app.Quit()
        }
    })

    deleteAccountButton := widget.NewButton("Delete Account", func() {
        confirmDialog := dialog.NewConfirm("Delete Account", "Are you sure you want to delete your account?", func(confirm bool) {
            if confirm {
                err := xmppfunctions.RemoveAccount(handler)
                if err != nil {
                    log.Printf("Account removal failed: %v", err)
                    dialog.ShowError(err, settingsWindow)
                    app.Quit()
                } else {
                    log.Println("Account removed successfully")
                    settingsWindow.Close()
                    app.Quit()
                }
            }
        }, settingsWindow)
        confirmDialog.SetDismissText("Cancel")
        confirmDialog.Show()
    })

    // Button to change the presence
    changePresenceButton := widget.NewButton("Change Presence", func() {
        ChangePresenceWindow(app, handler)
    })

    // Set the content of the settings window
    settingsWindow.SetContent(container.NewVBox(
        widget.NewLabel("User Settings"),
        routingOptions,             // Routing algorithm selection
        changePresenceButton,
        logoutButton,
        deleteAccountButton,
    ))

    settingsWindow.Resize(fyne.NewSize(300, 300))
    settingsWindow.Show()
}


func CloseAllWindows(app fyne.App) {
    for _, window := range app.Driver().AllWindows() {
        window.Close()
    }
}


func ShowCreateAccountDialog(app fyne.App, parent fyne.Window) {
    serverEntry := widget.NewEntry()
    serverEntry.SetPlaceHolder("Server (e.g., alumchat.lol:5222)")

    usernameEntry := widget.NewEntry()
    usernameEntry.SetPlaceHolder("Desired Username")

    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Desired Password")

    errorLabel := widget.NewLabel("")

    dialogWindow := app.NewWindow("Create Account")

    confirmButton := widget.NewButton("Create Account", func() {
        server := serverEntry.Text
        username := usernameEntry.Text
        password := passwordEntry.Text

        hostPort := strings.Split(server, ":")
        if len(hostPort) != 2 {
            errorLabel.SetText("Invalid server format")
            return
        }

        err := xmppfunctions.CreateUser(hostPort[0], hostPort[1], username, password)
        if err != nil {
            log.Printf("Account creation failed: %v", err)
            errorLabel.SetText(fmt.Sprintf("Error: %v", err))
            return
        }

        errorLabel.SetText("Account created successfully!")
        log.Println("Account created successfully")
        dialogWindow.Close() // Close the account creation window on success
    })

    content := container.NewVBox(
        widget.NewLabel("Create a New XMPP Account"),
        serverEntry,
        usernameEntry,
        passwordEntry,
        errorLabel,
        confirmButton,
    )

    dialogWindow.SetContent(content)
    dialogWindow.Resize(fyne.NewSize(300, 200))
    dialogWindow.Show()
}


func ChangePresenceWindow(app fyne.App, handler *xmpp.XMPPHandler) {
    presenceWindow := app.NewWindow("Change Presence")

    // Create a drop-down for the presence type
    presenceTypes := []string{"chat", "away", "dnd", "xa"}
    presenceTypeSelect := widget.NewSelect(presenceTypes, func(value string) {
        log.Println("Selected presence type:", value)
    })
    presenceTypeSelect.SetSelected("chat") // Default selection

    // Create an entry for the custom status message
    statusEntry := widget.NewEntry()
    statusEntry.SetPlaceHolder("Enter your status message")

    // Create a button to apply the changes
    applyButton := widget.NewButton("Apply", func() {
        selectedType := presenceTypeSelect.Selected
        statusMessage := statusEntry.Text

        // Send the presence update to the XMPP server
        err := handler.SendPresence(selectedType, statusMessage)
        if err != nil {
            log.Printf("Failed to change presence: %v", err)
            dialog.ShowError(err, presenceWindow)
        } else {
            log.Printf("Presence changed to: %s - %s", selectedType, statusMessage)
            presenceWindow.Close() // Close the window after applying the changes
        }
    })

    presenceWindow.SetContent(container.NewVBox(
        widget.NewLabel("Change Your Presence"),
        presenceTypeSelect,
        statusEntry,
        applyButton,
    ))

    presenceWindow.Resize(fyne.NewSize(300, 200))
    presenceWindow.Show()
}


func ShowContactDetailsWindow(app fyne.App, handler *xmpp.XMPPHandler, recipient xmppfunctions.Contact) {
    detailsWindow := app.NewWindow("Contact Details - " + recipient.JID)



    // if contact == nil {
    //     detailsWindow.SetContent(widget.NewLabel("Contact details not found."))
    // } else {
        details := container.NewVBox(
            widget.NewLabel("JID: " + recipient.JID),
            widget.NewLabel("Name: " + recipient.Name),
            widget.NewLabel("Subscription: " + recipient.Subscription),
            widget.NewLabel("Status: " + recipient.Status),
            widget.NewLabel("Presence: " + recipient.Presence),
        )

        detailsWindow.SetContent(details)
    // }

    detailsWindow.Resize(fyne.NewSize(300, 200))
    detailsWindow.Show()
}


func ShowNodesWindow(app fyne.App, handler *xmpp.XMPPHandler, routingTable routing.Network) {
    

    var nodes []string
    userNode := "" // The node that corresponds to the logged-in user

    // Iterate over the routing table to populate the nodes slice
    for nodeName, node := range routingTable.Nodes {
        nodes = append(nodes, nodeName)
        if node.Username == (handler.Username +"@"+ strings.Split(handler.Server, ":")[0]) {
            userNode = nodeName // Identify the user's node
        }
    }
    nodesWindow := app.NewWindow("Network Nodes - Node "+ userNode)
    // Function to refresh the node list
    refreshNodeList := func(nodeList *widget.List) {
        nodeList.Length = func() int {
            return len(nodes)
        }
        nodeList.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
            nodeName := nodes[i]
            queuedMessages := 0
            
            if node, exists := routingTable.Nodes[nodeName]; exists {
                jid := node.Username
                if msgs, ok := handler.MessageQueue[jid]; ok {
                    queuedMessages = len(msgs)
                }
            }
            displayText := nodeName
            if nodeName == userNode {
                displayText += " (You)"
            } else {
                if queuedMessages > 0{
                    displayText = fmt.Sprintf("%s (%d)", displayText, queuedMessages)
                }
            }
            o.(*widget.Label).SetText(displayText)
        }
        nodeList.Refresh()
    }

    nodeList := widget.NewList(
        func() int {
            return len(nodes)
        },
        func() fyne.CanvasObject {
            return widget.NewLabel("")
        },
        func(i widget.ListItemID, o fyne.CanvasObject) {},
    )

    refreshNodeList(nodeList)

    nodeList.OnSelected = func(id widget.ListItemID) {
        if id >= 0 && id < len(nodes) {
            selectedNode := nodes[id]
            selectedNodeDetails := routingTable.Nodes[selectedNode]

            if selectedNodeDetails != nil && selectedNodeDetails.Username != handler.Username {
                // Open the chat window with the selected node's corresponding user
                chatWindow := ShowChatWindow(app, handler, selectedNodeDetails.Username, xmppfunctions.Contact{
                    JID: selectedNodeDetails.Username,
                    Name: selectedNode, // Use node name as contact name
                    Status: selectedNodeDetails.ID, // You can add more details if available
                }, routingTable,userNode)
                handler.ChatWindows[selectedNodeDetails.Username] = chatWindow
            } else {
                log.Println("Cannot open a chat with yourself or an invalid node.")
            }
            nodeList.Unselect(id) // Unselect the node after showing the chat window
        } else {
            log.Printf("Invalid selection: %d", id)
        }
    }

    refreshButton := widget.NewButton("Refresh Nodes", func() {
        refreshNodeList(nodeList)
    })

    settingsButton := widget.NewButton("Settings", func() {
        ShowUserSettingsWindow(app, handler, &routingTable)
    })

    nodesWindow.SetContent(
        container.NewBorder(
            container.NewVBox(settingsButton,widget.NewLabel("Network Nodes"), refreshButton),
            nil, nil, nil,
            container.NewVScroll(nodeList),
        ),
    )

    nodesWindow.Resize(fyne.NewSize(300, 400))
    nodesWindow.Show()

    go func() {
        for msg := range handler.MessageChan {
            fmt.Println("Processing incoming message: ", msg)
            parsedMessage, err := ParseXMPPMessageBody(msg)
            if err != nil{
                fmt.Println("Error parsing xmpp message: ", err)
            }else{
                ProcessMessage(handler, userNode,parsedMessage,routingTable,msg)
            }
            
        }
    }()

    handler.ListenForIncomingStanzas()   
}

 


