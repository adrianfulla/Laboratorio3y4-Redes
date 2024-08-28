package xmppfunctions

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"strings"
    // "time"
	"sync"

	"github.com/adrianfulla/Proyecto1-Redes/server/xmpp"
)

// CreateUser creates a new account on the XMPP server.
func CreateUser(domain,port, username, password string) error {
    handler := &xmpp.XMPPHandler{
        Server:   domain +":"+port,
        Username: username,
        Password: password,
    }

    conn, err := xmpp.NewXMPPConnection(domain, port, false)
    if err != nil {
        return err
    }
    handler.Conn = conn

    if err := handler.Conn.StartStream(""); err != nil {
        return err
    }

    // If the user does not exist, create it
    if err := xmpp.CreateUser(handler.Conn, username, password); err != nil {
        log.Println("User creation failed or user already exists, proceeding with login...")
        return err
    }

    return nil
}

// Login authenticates a user and returns an XMPPHandler.
func Login(domain,port, username, password string) (*xmpp.XMPPHandler, error) {
    handler, err := xmpp.NewXMPPHandler(domain,port, username, password)
    if err != nil {
        return nil, err
    }
    // handler.SendPresence("presence", "Online")
    return handler, nil
}

// Logout closes the XMPP connection gracefully.
func Logout(handler *xmpp.XMPPHandler) error {
    if handler == nil || handler.Conn == nil {
        return errors.New("invalid handler")
    }
    defer handler.Conn.Close()
    // You may send unavailable presence before logging out
    return handler.SendPresence("unavailable", "Logging out")
}

// RemoveAccount removes a user account from the XMPP server.
func RemoveAccount(handler *xmpp.XMPPHandler) error {
    if handler == nil || handler.Conn == nil {
        return errors.New("invalid handler")
    }

    iqID := "remove1"
    removeRequest := `<iq type="set" id="` + iqID + `"><query xmlns="jabber:iq:register"><remove/></query></iq>`
    _, err := handler.Conn.Conn.Write([]byte(removeRequest))
    if err != nil {
        return fmt.Errorf("failed to send remove request: %v", err)
    }

    // Wait for the response
    buffer := make([]byte, 4096)
    n, err := handler.Conn.Conn.Read(buffer)
    if err != nil {
        return fmt.Errorf("error reading remove account response: %v", err)
    }

    response := string(buffer[:n])
    if strings.Contains(response, "<iq type=\"result\"") {
        log.Println("Account removed successfully")
        return nil
    }

    return fmt.Errorf("failed to remove account: %s", response)
}

// GetContacts retrieves the user's roster (contact list).
func GetContacts(handler *xmpp.XMPPHandler) ([]Contact, error) {
    iqID := "getRoster1"
    rosterRequest := `<iq type="get" id="` + iqID + `"><query xmlns="jabber:iq:roster"/></iq>`

    var iq IQ


    _, err := handler.Conn.Conn.Write([]byte(rosterRequest))
    if err != nil {
        return nil, fmt.Errorf("failed to send roster request: %v", err)
    }

    buffer := make([]byte, 4096)
    n, err := handler.Conn.Conn.Read(buffer)
    if err != nil {
        return nil, fmt.Errorf("error reading roster response: %v", err)
    }
    response := string(buffer[:n])
    fmt.Printf("Obtained response: %s\n", response)
    err = xml.Unmarshal([]byte(response), &iq)

    contacts := []Contact{}
    var wg sync.WaitGroup

    for _, item := range iq.Query.Items {
        wg.Add(1)
        contact := Contact{
            JID:          item.JID,
            Name:         item.Name,
            Subscription: item.Subscription,
            Presence:     "unavailable", // Default to unavailable
            Status:       "Offline",
        }
        log.Printf(contact.JID)
        // Check if there is a presence for this contact in the PresenceStack
        if presence, found := handler.PresenceStack[contact.JID]; found {
            
            contact.Presence = presence.Show
            if (presence.HasStatus()){
                contact.Status = presence.Status
            }else if (presence.IsAvailable()) {
                contact.Status = "Online"
            } else{
                contact.Status = "Offline"
            }
        }

        contacts = append(contacts, contact)
    }

    return contacts, nil
}

func CheckContacts(handler *xmpp.XMPPHandler, contacts []Contact) ([]Contact){
    newContacts := []Contact{}
    for _,contact := range contacts{
        if presence, found := handler.PresenceStack[contact.JID]; found {
            contact.Presence = presence.Show
            if (presence.HasStatus()){
                contact.Status = presence.Status
            }else if (presence.IsAvailable()) {
                contact.Status = "Online"
            } else{
                contact.Status = "Offline"
            }
        }
        newContacts = append(newContacts, contact)
    }
    return newContacts
}

// AddContact adds a new contact to the user's roster.
func AddContact(handler *xmpp.XMPPHandler, jid string) error {
    // Send a presence subscription request to the new contact
    subscriptionRequest := fmt.Sprintf(
        `<presence to='%s' type='subscribe'/>`,
        jid,
    )

    _, err := handler.Conn.Conn.Write([]byte(subscriptionRequest))
    if err != nil {
        return fmt.Errorf("failed to send subscription request: %v", err)
    }

    return nil
}


// GetContactDetails retrieves details about a specific contact.
func GetContactDetails(handler *xmpp.XMPPHandler, contactJID string) (ContactDetails, error) {
    iqID := "v1"
    vCardRequest := fmt.Sprintf(`<iq from='%s' type='get' id='%s'><vCard xmlns='vcard-temp'/></iq>`, contactJID, iqID)

    // Send the vCard request
    fmt.Println(vCardRequest)
    _, err := handler.Conn.Conn.Write([]byte(vCardRequest))
    if err != nil {
        return ContactDetails{}, fmt.Errorf("failed to send vCard request: %v", err)
    }

    // Wait for the response
    // buffer := make([]byte, 4096)
    // n, err := handler.Conn.Conn.Read(buffer)
    // if err != nil {
    //     return ContactDetails{}, fmt.Errorf("error reading vCard response: %v", err)
    // }

    // response 
    // fmt.Printf("vCard Response: %s\n", response)

    // // Parse the vCard response
    // var iq xmpp.IQ
    // err = xml.Unmarshal([]byte(response), &iq)
    // if err != nil {
    //     return ContactDetails{}, fmt.Errorf("failed to parse vCard response: %v", err)
    // }

    // Extract the vCard info from the IQ response
    var vCard vCardQuery
    // err = xml.Unmarshal(iq.Query, &vCard)
    // if err != nil {
    //     return ContactDetails{}, fmt.Errorf("failed to parse vCard query: %v", err)
    // }

    // Build the ContactDetails struct
    details := ContactDetails{
        JID:       contactJID,
        Name:      vCard.FullName,
        VCardInfo: fmt.Sprintf("Nickname: %s\nEmail: %s", vCard.Nickname, vCard.Email),
    }

    return details, nil
}

// SendMessage sends a one-to-one message to a specific user.
func SendMessage(handler *xmpp.XMPPHandler, to, message string) error {
    return handler.SendMessage(to, message)
}

// JoinGroupChat allows the user to join a multi-user chat room.
func JoinGroupChat(handler *xmpp.XMPPHandler, roomJID, nickname string) error {
    presence := fmt.Sprintf(
        `<presence to='%s/%s'><x xmlns='http://jabber.org/protocol/muc'/></presence>`,
        roomJID, nickname,
    )
    _, err := handler.Conn.Conn.Write([]byte(presence))
    return err
}

// SendNotification sends a notification to a user or group.
func SendNotification(handler *xmpp.XMPPHandler, to, notification string) error {
    message := fmt.Sprintf(
        `<message to='%s' type='headline'><body>%s</body></message>`,
        to, notification,
    )
    _, err := handler.Conn.Conn.Write([]byte(message))
    return err
}

// SendFile sends a file to a specific contact.
func SendFile(handler *xmpp.XMPPHandler, to, filePath string) error {
    // Implement file transfer using XMPP's file transfer protocols
    return nil
}

// ReceiveMessages handles incoming messages and stanzas.
func ReceiveMessages(handler *xmpp.XMPPHandler) (string, error) {
    // The HandleIncomingStanzas now returns on errors, so we use it in a loop.
    for {
        err := handler.HandleIncomingStanzas()
        if err != nil {
            return "", err
        }
    }
}



type Contact struct {
    JID    string
    Name   string
    Subscription string
    Presence string
    Status string
}

type ContactDetails struct {
    JID       string
    Name      string
    VCardInfo string
}

type RosterQuery struct {
	XMLName xml.Name    `xml:"query"`
	Items   []RosterItem `xml:"item"`
}

type RosterItem struct {
	JID          string `xml:"jid,attr"`
	Subscription string `xml:"subscription,attr"`
	Name         string `xml:"name,attr,omitempty"`
}

type IQ struct {
	XMLName xml.Name   `xml:"iq"`
	Type    string     `xml:"type,attr"`
	ID      string     `xml:"id,attr"`
	To      string     `xml:"to,attr,omitempty"`
	Query   RosterQuery `xml:"query"`
}


type vCardQuery struct {
    XMLName xml.Name `xml:"vCard"`
    FullName string  `xml:"FN,omitempty"`
    Nickname string  `xml:"NICKNAME,omitempty"`
    Email    string  `xml:"EMAIL>USERID,omitempty"`
}
