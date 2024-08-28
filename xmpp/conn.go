package xmpp

import (
	"crypto/tls"
	"net"
	"time"
	"fmt"
	"log"
	"errors"
	"strings"
)

type XMPPConnection struct {
    Conn net.Conn
	Domain string
}

func NewXMPPConnection(domain string,port string, useTLS bool) (*XMPPConnection, error) {
    var conn net.Conn
    var err error
	address := domain+":"+port

	dialer := &net.Dialer{
        Timeout: 5 * time.Second,
    }

    if useTLS {
		// TLS config with InsecureSkipVerify to bypass certificate verification
        tlsConfig := &tls.Config{
            InsecureSkipVerify: true, // Disable certificate verification
            ServerName:         "alumchat.lol",
        }

        // Directly establish a TLS connection, effectively skipping STARTTLS
        conn, err = tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
        if err != nil {
            return nil, fmt.Errorf("TLS connection failed: %w", err)
        }
    } else {
        conn, err = net.Dial("tcp", address)
    }

    if err != nil {
        return nil, err
    }

    return &XMPPConnection{Conn: conn, Domain: domain}, nil
}

func (xc *XMPPConnection) Close() error {
    return xc.Conn.Close()
}


// StartTLS sends the STARTTLS command to the server and upgrades the connection to TLS.
func StartTLS(conn *XMPPConnection) error {
    startTLS := "<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>"
    if _, err := conn.Conn.Write([]byte(startTLS)); err != nil {
        return fmt.Errorf("failed to send STARTTLS: %v", err)
    }

    buffer := make([]byte, 4096)
    n, err := conn.Conn.Read(buffer)
    if err != nil {
        return fmt.Errorf("failed to read STARTTLS response: %v", err)
    }

    response := string(buffer[:n])
    log.Printf("Received STARTTLS response: %s\n", response)

    if strings.Contains(response, "<proceed") {
        log.Println("Proceeding with TLS handshake...")
        tlsConn := tls.Client(conn.Conn, &tls.Config{
            InsecureSkipVerify: true, // Disable verification for testing, not recommended for production
        })
        conn.Conn = tlsConn
        if err := tlsConn.Handshake(); err != nil {
            return fmt.Errorf("TLS handshake failed: %v", err)
        }
        log.Println("TLS handshake successful")
        return nil
    }

    return errors.New("failed to initiate STARTTLS")
}

// sendStanza sends a stanza over the XMPP connection and logs it.
func sendStanza(conn *XMPPConnection, stanza Stanza) error {
    xml, err := stanza.ToXML()
    if err != nil {
        return fmt.Errorf("failed to marshal stanza to XML: %v", err)
    }
    
    // Log the outgoing XML
    log.Printf("Sending XML: %s", xml)

    _, err = conn.Conn.Write([]byte(xml))
    return err
}