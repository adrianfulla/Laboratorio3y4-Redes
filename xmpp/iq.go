package xmpp

import (
    "encoding/xml"
    "errors"
    "fmt"
    "log"
    "strings"
)

type RegisterRequest struct {
    XMLName  xml.Name `xml:"query"`
    XMLNS    string   `xml:"xmlns,attr"`
    Username string   `xml:"username"`
    Password string   `xml:"password"`
}

type RawXML []byte

type IQ struct {
    XMLName xml.Name `xml:"iq"`
    From    string   `xml:"from,attr,omitempty"`
    To      string   `xml:"to,attr,omitempty"`
    Type    string   `xml:"type,attr"`
    ID      string   `xml:"id,attr"`
    Query   interface{} `xml:",omitempty"`
}

type IQItem struct {
    JID          string `xml:"jid,attr"`
    Name         string `xml:"name,attr,omitempty"`
    Subscription string `xml:"subscription,attr"`
}


func NewIQ(iqType, iqID string) *IQ {
    return &IQ{
        Type: iqType,
        ID:   iqID,
    }
}


func (iq *IQ) SetQuery(query interface{}) {
    iq.Query = query
}

func (iq *IQ) ToXML() (string, error) {
    output, err := xml.Marshal(iq)
    if err != nil {
        return "", err
    }
    return string(output), nil
}

func CreateUser(conn *XMPPConnection, username, password string) error {
    // Prepare the registration request
    iqID := "register1"
    register := RegisterRequest{
        XMLNS:    "jabber:iq:register", // Set the correct namespace
        Username: username,
        Password: password,
    }

    iq := NewIQ("set", iqID)
    iq.SetQuery(register)
    
    if err := sendStanza(conn, iq); err != nil {
        return fmt.Errorf("failed to send registration request: %v", err)
    }

    // Wait for the response and handle it
    buffer := make([]byte, 4096)
    n, err := conn.Conn.Read(buffer)
    if err != nil {
        return fmt.Errorf("error reading registration response: %v", err)
    }

    response := string(buffer[:n])
    log.Printf("Received registration response: %s\n", response)

    // Handle <stream:features> response and check if it includes registration result
    if strings.Contains(response, "<stream:features>") {
        log.Println("Stream features received. Continuing to monitor for registration result...")
        // Check if the response also contains an IQ result/error directly after <stream:features>
        if strings.Contains(response, "<iq type=\"result\"") {
            log.Println("User created successfully")
            return nil
        }
        if strings.Contains(response, "<iq type=\"error\"") {
            if strings.Contains(response, "<conflict") {
                return fmt.Errorf("user already exists")
            }
            return fmt.Errorf("failed to create user: %s", extractErrorMessage(response))
        }
        // If not, read again to get the IQ result/error
        n, err = conn.Conn.Read(buffer)
        if err != nil {
            return fmt.Errorf("error reading final registration response: %v", err)
        }
        response = string(buffer[:n])
        log.Printf("Received final registration response: %s\n", response)
    }

    // Handle successful user creation
    if strings.Contains(response, "<iq type=\"result\"") {
        log.Println("User created successfully")
        return nil
    }

    // Handle error during user creation
    if strings.Contains(response, "<iq type=\"error\"") {
        if strings.Contains(response, "<conflict") {
            return fmt.Errorf("user already exists")
        }
        return fmt.Errorf("failed to create user: %s", extractErrorMessage(response))
    }

    return errors.New("unexpected registration response")
}

// Helper function to extract error message from the server response
func extractErrorMessage(response string) string {
    start := strings.Index(response, "<error")
    if start == -1 {
        return "unknown error"
    }
    end := strings.Index(response[start:], "</error>")
    if end == -1 {
        end = len(response)
    } else {
        end += start + len("</error>")
    }
    return response[start:end]
}


func BindResource(conn *XMPPConnection) error {
    // Resource binding request
    iqStanza := `<iq type="set" id="bind_1">
                    <bind xmlns="urn:ietf:params:xml:ns:xmpp-bind">
                        <resource>mainbinding</resource>
                    </bind>
                 </iq>`

    // Send the IQ stanza for resource binding
    _, err := conn.Conn.Write([]byte(iqStanza))
    if err != nil {
        return fmt.Errorf("failed to send resource binding request: %v", err)
    }

    // Wait for the response
    buffer := make([]byte, 4096)
    n, err := conn.Conn.Read(buffer)
    if err != nil {
        return fmt.Errorf("error reading resource binding response: %v", err)
    }

    response := string(buffer[:n])
    log.Printf("Received resource binding response: %s\n", response)

    if strings.Contains(response, "<iq type=\"result\"") {
        log.Println("Resource binding successful")
        return nil
    }

    return fmt.Errorf("resource binding failed or unexpected response")
}