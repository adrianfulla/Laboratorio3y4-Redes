package xmpp

import (
    "encoding/xml"
)

// Presence represents an XMPP presence stanza.
type Presence struct {
    XMLName  xml.Name `xml:"presence"`
    From     string   `xml:"from,attr,omitempty"`
    To       string   `xml:"to,attr,omitempty"`
    Type     string   `xml:"type,attr,omitempty"` // "available", "unavailable", "subscribe", etc.
    Show     string   `xml:"show,omitempty"`      // "chat", "away", "dnd", "xa" (extended away)
    Status   string   `xml:"status,omitempty"`    // User-defined status message
    Priority int      `xml:"priority,omitempty"`  // Priority level (-128 to +127)
}

// NewPresence creates a new presence stanza with the specified parameters.
func NewPresence(to, presenceType, show, status string, priority int) *Presence {
    return &Presence{
        To:       to,
        Type:     presenceType,
        Show:     show,
        Status:   status,
        Priority: priority,
    }
}

// ToXML converts the Presence struct to an XML string.
func (p *Presence) ToXML() (string, error) {
    output, err := xml.Marshal(p)
    if err != nil {
        return "", err
    }
    return string(output), nil
}

// ParsePresence parses an XML string into a Presence struct.
func ParsePresence(data []byte) (*Presence, error) {
    var presence Presence
    if err := xml.Unmarshal(data, &presence); err != nil {
        return nil, err
    }
    return &presence, nil
}

// Additional utility functions can be added below

// IsAvailable checks if the presence type is "available".
func (p *Presence) IsAvailable() bool {
    return p.Type == "" || p.Type == "available"
}

func (p *Presence) HasStatus() bool {
    return p.Status != ""
}

// IsUnavailable checks if the presence type is "unavailable".
func (p *Presence) IsUnavailable() bool {
    return p.Type == "unavailable"
}

// IsSubscriptionRequest checks if the presence type is "subscribe" or "subscribed".
func (p *Presence) IsSubscriptionRequest() bool {
    return p.Type == "subscribe" || p.Type == "subscribed"
}
