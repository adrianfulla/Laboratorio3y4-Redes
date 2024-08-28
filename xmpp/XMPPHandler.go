package xmpp

import (
	"encoding/xml"
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
    "fyne.io/fyne/v2/widget"
    "fyne.io/fyne/v2/dialog"
)

type XMPPHandler struct {
    Conn     *XMPPConnection
    Server   string
    Username string
    Password string
    ChatWindows map[string]*ChatWindow
    MessageChan chan *Message
    MessageQueue map[string][]*Message
    PresenceStack map[string]*Presence
    VCardStack map[string]*IQ
}
type ChatWindow struct {
    Window       fyne.Window
    ChatContent  *fyne.Container
    Handler      *XMPPHandler
    Recipient    string
}

func (cw *ChatWindow) AddMessage(msg *Message) {
    // Create a new label for the incoming message and add it to the chat content
    messageLabel := widget.NewLabel(fmt.Sprintf("%s: %s", strings.Split(msg.From,"/")[0], msg.Body))
    cw.ChatContent.Add(messageLabel)
    
    // Refresh the window to display the new message
    cw.Window.Content().Refresh()
}

func NewXMPPHandler(domain, port, username, password string) (*XMPPHandler, error) {
    handler := &XMPPHandler{
        Server:   domain +":"+port,
        Username: username,
        Password: password,
        ChatWindows: make(map[string]*ChatWindow),
        MessageChan:  make(chan *Message, 100),
        MessageQueue: make(map[string][]*Message),
        PresenceStack: make(map[string]*Presence),
        VCardStack: map[string]*IQ{},
    }

    conn, err := NewXMPPConnection(domain, port, false)
    if err != nil {
        return nil, err
    }
    handler.Conn = conn

    if err := handler.Conn.StartStream(""); err != nil {
        return nil, err
    }

    // Authenticate
    if err := Authenticate(handler.Conn, username, password); err != nil {
        return nil, err
    }

    // Bind Resource
    if err := BindResource(handler.Conn); err != nil {
        return nil, err
    }

    return handler, nil
}


// SendPresence sends a presence stanza to update the user's availability status.
func (h *XMPPHandler) SendPresence(presenceType, status string) error {
    presence := fmt.Sprintf(
        `<presence><show>%s</show><status>%s</status></presence>`,
        presenceType, status,
    )
    _, err := h.Conn.Conn.Write([]byte(presence))
    if err != nil {
        log.Printf("Failed to send presence: %v", err)
        return err
    }
    log.Println("Presence sent successfully")
    return nil
}

// SendMessage sends a message stanza to the specified recipient.
func (h *XMPPHandler) SendMessage(to, message string) error {
    msg := fmt.Sprintf(
        `<message to='%s' type='chat'><body>%s</body></message>`,
        to, message,
    )
    _, err := h.Conn.Conn.Write([]byte(msg))
    if err != nil {
        log.Printf("Failed to send message: %v", err)
        return err
    }
    log.Println("Message sent successfully")
    return nil
}

func (h *XMPPHandler) HandleIncomingStanzas() error {
	decoder := xml.NewDecoder(h.Conn.Conn)

	for {
		tok, err := decoder.Token()
		if err != nil {
			log.Printf("Failed to read stanza: %v", err)
			return err
		}
        
		switch se := tok.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "message":
				var msg Message
                if err := decoder.DecodeElement(&msg, &se); err != nil {
                    log.Printf("Failed to parse message: %v", err)
                    continue
                }
                h.DispatchMessage(&msg)

			case "presence":
				var pres Presence
				if err := decoder.DecodeElement(&pres, &se); err != nil {
					log.Printf("Failed to parse presence: %v", err)
					continue
				}
				h.handlePresence(&pres)

			case "iq":
				var iq IQ
                log.Printf("Obtained IQ: %s", se)
				if err := decoder.DecodeElement(&iq, &se); err != nil {
					log.Printf("Failed to parse IQ: %v", err)
					continue
				}
                h.handleIQ(&iq)
                
				

			default:
				log.Printf("Unhandled stanza type: %s", se.Name.Local)
			}
		}
	}
}

func (h *XMPPHandler) handlePresence(pres *Presence) {
	jid := strings.Split(pres.From, "/")[0]
    log.Printf("Presence from %s: %s|%s|%s", jid, pres.Status, pres.Show, pres.Type)

    switch pres.Type{
    case "subscribe":
        fyne.CurrentApp().SendNotification(&fyne.Notification{
            Title:   "Subscription Request",
            Content: fmt.Sprintf("%s wants to subscribe to your presence", pres.From),
        })
        h.PromptSubscriptionRequest(pres.From)
    
    case "subscribed":
        // Your subscription request was accepted
        fyne.CurrentApp().SendNotification(&fyne.Notification{
            Title:   "Subscription Accepted",
            Content: fmt.Sprintf("%s accepted your subscription request", pres.From),
        })
    case "unsubscribe":
        // Someone wants to unsubscribe from your presence
        fyne.CurrentApp().SendNotification(&fyne.Notification{
            Title:   "Unsubscription Request",
            Content: fmt.Sprintf("%s wants to unsubscribe from your presence", pres.From),
        })
    case "unsubscribed":
        // Your subscription request was rejected or someone unsubscribed
        fyne.CurrentApp().SendNotification(&fyne.Notification{
            Title:   "Subscription Rejected",
            Content: fmt.Sprintf("%s has rejected your subscription or unsubscribed", pres.From),
        })
    default:
        if pres.Type != "error"{
            h.PresenceStack[jid] = pres
        }
    }
}

func (h *XMPPHandler) PromptSubscriptionRequest(from string) {
    confirmDialog := dialog.NewConfirm("Subscription Request", fmt.Sprintf("%s wants to subscribe to your presence. Do you accept?", from), func(confirm bool) {
        if confirm {
            // Send 'subscribed' presence to accept the subscription
            presence := NewPresence(from, "subscribed", "", "", 0)
            presenceXML, err := presence.ToXML()
            if err != nil {
                log.Printf("Failed to marshal subscription acceptance: %v", err)
                return
            }
            _, err = h.Conn.Conn.Write([]byte(presenceXML))
            if err != nil {
                log.Printf("Failed to send subscription acceptance: %v", err)
            } else {
                log.Printf("Subscription accepted for %s", from)
            }
        } else {
            // Send 'unsubscribed' presence to reject the subscription
            presence := NewPresence(from, "unsubscribed", "", "", 0)
            presenceXML, err := presence.ToXML()
            if err != nil {
                log.Printf("Failed to marshal subscription rejection: %v", err)
                return
            }
            _, err = h.Conn.Conn.Write([]byte(presenceXML))
            if err != nil {
                log.Printf("Failed to send subscription rejection: %v", err)
            } else {
                log.Printf("Subscription rejected for %s", from)
            }
        }
    }, fyne.CurrentApp().Driver().AllWindows()[0]) // Assuming you want to show the dialog on the main window

    confirmDialog.SetDismissText("Ignore")
    confirmDialog.Show()
}


func (h *XMPPHandler) handleIQ(iq *IQ) {
    // Check if the IQ has a known type but no specific query body
    if iq.Type == "get" || iq.Type == "set" {
        // Handle specific IQ requests, like version or ping
        if iq.Query == nil {
            log.Printf("Received IQ request without specific query from %s", iq.From)

            // Determine the response type based on common XMPP queries
            switch iq.ID {
            // Add cases for known IQ requests based on the ID or other attributes
            default:
                // If we don't recognize the specific IQ request, we can send a basic result
                h.sendIQResult(iq)
            }
        } else {
            // If there's a query body, handle it accordingly
            switch query := iq.Query.(type) {
            case struct{ XMLName xml.Name }:
                switch query.XMLName.Space {
                case "jabber:iq:version":
                    h.handleVersionQuery(iq)
                // case "jabber:iq:roster":
                //     h.handleRosterQuery(iq)
                default:
                    log.Printf("Unhandled IQ namespace: %s", query.XMLName.Space)
                }
            default:
                log.Printf("Unhandled type in IQ query: %T", iq.Query)
            }
        }
    } else if iq.Type == "result"{
        log.Printf("Received result IQ from %s: %s|%s|%s", iq.From, iq.To, iq.Type, iq.Query)
        switch iq.ID  {
        case "v1":
            h.VCardStack[iq.From] = iq
        
        default: 
        log.Printf("Unhandled type in IQ ID: %T", iq.ID)
        }
        
    }
}

func (h *XMPPHandler) sendIQResult(iq *IQ) {
    response := IQ{
        XMLName: xml.Name{Local: "iq"},
        Type:    "result",
        ID:      iq.ID,
        To:      iq.From,
    }

    // Convert to XML and send the response
    xmlResponse, err := response.ToXML()
    if err != nil {
        log.Printf("Failed to marshal IQ response: %v", err)
        return
    }

    _, err = h.Conn.Conn.Write([]byte(xmlResponse))
    if err != nil {
        log.Printf("Failed to send IQ response: %v", err)
    } else {
        log.Printf("Sent IQ response to %s", iq.From)
    }
}

func (h *XMPPHandler) handleVersionQuery(iq *IQ) {
    log.Printf("Received version query from %s", iq.From)

    response := fmt.Sprintf(
        `<iq type='result' id='%s' to='%s'>
            <query xmlns='jabber:iq:version'>
                <name>XMPP Client</name>
                <version>1.0</version>
                <os>Go</os>
            </query>
        </iq>`, iq.ID, iq.From)

    _, err := h.Conn.Conn.Write([]byte(response))
    if err != nil {
        log.Printf("Failed to send IQ response: %v", err)
    } else {
        log.Printf("Sent IQ response to %s", iq.From)
    }
}

func (h *XMPPHandler) RequestOfflineMessages() error {
    iq := IQ{
        XMLName: xml.Name{Local: "iq"},
        Type:    "get",
        ID:      "offline1",
        Query: struct {
            XMLName xml.Name `xml:"offline"`
        }{
            XMLName: xml.Name{Local: "query", Space: "jabber:iq:offline"},
        },
    }

    iqXML, err := xml.Marshal(iq)
    if err != nil {
        return fmt.Errorf("failed to marshal offline message request: %v", err)
    }

    _, err = h.Conn.Conn.Write(iqXML)
    if err != nil {
        return fmt.Errorf("failed to send offline message request: %v", err)
    }

    log.Println("Offline message request sent successfully")
    return nil
}


// WaitForShutdown keeps the connection alive until shutdown is requested.
func (h *XMPPHandler) WaitForShutdown() {
    log.Println("Waiting for shutdown signal...")
    // Implementation could wait on a signal or just block until interrupted
    select {}
}


func (h *XMPPHandler) LoginAndFetchMessages(status string) error {
    // Send presence to indicate the client is online
    if err := h.SendPresence("presence",status); err != nil {
        return err
    }

    // Request offline messages (if the server requires this)
    if err := h.RequestOfflineMessages(); err != nil {
        return err
    }

    // Start listening for incoming messages
    go func() {
        if err := h.HandleIncomingStanzas(); err != nil {
            log.Printf("Error handling incoming stanzas: %v", err)
        }
    }()

    return nil
}


func (h *XMPPHandler) ListenForIncomingStanzas() {
    h.SendPresence("presence", "Online")
    go func() {
        for {
            err := h.HandleIncomingStanzas()
            if err != nil {
                log.Printf("Error handling stanzas: %v", err)
                continue
            }
        }
    }()
}

func (h *XMPPHandler) DispatchMessage(msg *Message) {
    recipient := strings.Split(msg.From, "/")[0]

    if chatWindow, ok := h.ChatWindows[recipient]; ok && chatWindow != nil {
        chatWindow.AddMessage(msg)
        fyne.CurrentApp().SendNotification(&fyne.Notification{
            Title:   "New Message",
            Content: fmt.Sprintf("%s: %s", recipient, msg.Body),
        })
    } else {
        if len(msg.Body) > 0 {
            log.Printf("No chat window open for %s, queueing message", recipient)
            h.MessageQueue[recipient] = append(h.MessageQueue[recipient], msg)
            
            // Send notification for each queued message
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "New Message",
                Content: fmt.Sprintf("%s: %s", recipient, msg.Body),
            })
        }
    }
}










