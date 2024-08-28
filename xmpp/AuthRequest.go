package xmpp

import (
    "encoding/xml"
    "encoding/base64"
    "errors"
    "fmt"
    "log"
    "strings"
)

type AuthRequest struct {
    XMLName   xml.Name `xml:"auth"`
    Mechanism string   `xml:"mechanism,attr"`
    Text      string   `xml:",chardata"`
}

func (a *AuthRequest) ToXML() (string, error) {
    output, err := xml.Marshal(a)
    if err != nil {
        return "", fmt.Errorf("failed to marshal AuthRequest: %v", err)
    }
    return string(output), nil
}

// Authenticate performs SASL PLAIN authentication with the XMPP server.
// func Authenticate(conn *XMPPConnection, username, password string) error {
//     // Perform STARTTLS if available
//     if err := StartTLS(conn); err != nil {
//         return fmt.Errorf("STARTTLS failed: %v", err)
//     }

//     // Reinitiate stream
//     if err := conn.StartStream(""); err != nil {
//         return fmt.Errorf("failed to start stream after STARTTLS: %v", err)
//     }

//     // Read initial response and check for stream features
//     buffer := make([]byte, 4096)
//     n, err := conn.Conn.Read(buffer)
//     if err != nil {
//         return fmt.Errorf("error reading initial response after STARTTLS: %v", err)
//     }

//     initialResponse := string(buffer[:n])
//     log.Printf("Received initial response after STARTTLS: %s\n", initialResponse)

//     if !strings.Contains(initialResponse, "<stream:features>") {
//         return errors.New("expected <stream:features> but did not receive it after STARTTLS")
//     }

//     // Perform SASL PLAIN authentication
//     authText := "\x00" + username + "\x00" + password
//     authBase64 := base64.StdEncoding.EncodeToString([]byte(authText))
//     auth := AuthRequest{
//         Mechanism: "PLAIN",
//         Text:      authBase64,
//     }

//     if err := sendStanza(conn, &auth); err != nil {
//         return fmt.Errorf("failed to send authentication request: %v", err)
//     }

//     // Wait for the response
//     n, err = conn.Conn.Read(buffer)
//     if err != nil {
//         return fmt.Errorf("error reading authentication response: %v", err)
//     }

//     response := string(buffer[:n])
//     log.Printf("Received authentication response: %s\n", response)

//     if strings.Contains(response, "<success") {
//         log.Println("Authentication successful")
//         return nil
//     }

//     return errors.New("authentication failed or unexpected response")
// }


// Authenticate performs SASL PLAIN authentication with the XMPP server.
func Authenticate(conn *XMPPConnection, username, password string) error {
    StartTLS_trys := 0
    var err error
    // Perform STARTTLS if available
    for StartTLS_trys < 5{
            err = StartTLS(conn)
            if err == nil{
                break
            }
            StartTLS_trys ++
    } 
    if err != nil{
        return fmt.Errorf("STARTTLS failed: %v", err)
    }

    // Reinitiate the stream after STARTTLS
    if err := conn.StartStream(""); err != nil {
        return fmt.Errorf("failed to start stream after STARTTLS: %v", err)
    }

    // Read and handle the initial <stream:features> response
    buffer := make([]byte, 4096)
    n, err := conn.Conn.Read(buffer)
    if err != nil {
        return fmt.Errorf("error reading initial response after STARTTLS: %v", err)
    }

    initialResponse := string(buffer[:n])
    log.Printf("Received initial response after STARTTLS: %s\n", initialResponse)

    if !strings.Contains(initialResponse, "<stream:features>") {
        return errors.New("expected <stream:features> but did not receive it after STARTTLS")
    }

 	authText := "\x00" + username + "\x00" + password
	// log.Printf("AuthText before Base64: %s", authText)
    authBase64 := base64.StdEncoding.EncodeToString([]byte(authText))
    authStanza := fmt.Sprintf(`<auth xmlns="urn:ietf:params:xml:ns:xmpp-sasl" mechanism="PLAIN">%s</auth>`, authBase64)


	// Log the outgoing authentication stanza
    log.Printf("Sending authentication stanza: %s", authStanza)

    // Send the authentication stanza
    _, err = conn.Conn.Write([]byte(authStanza))
    if err != nil {
        return fmt.Errorf("failed to send authentication request: %v", err)
    }

    // Wait for the response
    n, err = conn.Conn.Read(buffer)
    if err != nil {
        return fmt.Errorf("error reading authentication response: %v", err)
    }

    response := string(buffer[:n])
    log.Printf("Received authentication response: %s\n", response)

    // Check for successful authentication
    if strings.Contains(response, "<success") {
        log.Println("Authentication successful")
        return nil
    }

    // Handle authentication failure
    return errors.New("authentication failed or unexpected response")
}

