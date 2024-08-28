package xmpp

import (
    "encoding/xml"
)

// Message represents an XMPP message stanza.
type Message struct {
    XMLName xml.Name `xml:"message"`
    To      string   `xml:"to,attr,omitempty"`
    From    string   `xml:"from,attr,omitempty"`
    Type    string   `xml:"type,attr,omitempty"`
    Body    string   `xml:"body,omitempty"`
    Subject string   `xml:"subject,omitempty"`
    Thread  string   `xml:"thread,omitempty"`
}

// NewMessage creates a new message with the specified type, recipient, and body.
func NewMessage(to, msgType, body string) *Message {
    return &Message{
        To:   to,
        Type: msgType,
        Body: body,
    }
}

// ToXML converts the Message struct to an XML string.
func (m *Message) ToXML() (string, error) {
    output, err := xml.Marshal(m)
    if err != nil {
        return "", err
    }
    return string(output), nil
}

// ParseMessage parses an XML string into a Message struct.
func ParseMessage(data []byte) (*Message, error) {
    var msg Message
    if err := xml.Unmarshal(data, &msg); err != nil {
        return nil, err
    }
    return &msg, nil
}

// Additional utility functions can be added below

// IsChatMessage checks if the message is of type "chat".
func (m *Message) IsChatMessage() bool {
    return m.Type == "chat"
}

// IsGroupChatMessage checks if the message is of type "groupchat".
func (m *Message) IsGroupChatMessage() bool {
    return m.Type == "groupchat"
}

// IsErrorMessage checks if the message is of type "error".
func (m *Message) IsErrorMessage() bool {
    return m.Type == "error"
}
