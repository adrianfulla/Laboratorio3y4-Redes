package xmpp

import (
    "encoding/xml"
)

type Stanza interface {
    ToXML() (string, error)
}

func ParseStanza(data []byte) (Stanza, error) {
    var msg Message
    if err := xml.Unmarshal(data, &msg); err != nil {
        return nil, err
    }
    return &msg, nil
}
