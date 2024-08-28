package xmpp

import (
    "fmt"
    "io"
)

func (xc *XMPPConnection) StartStream(domain string) error {
    streamHeader := fmt.Sprintf("<?xml version='1.0'?><stream:stream to='%s' xmlns='jabber:client' xmlns:stream='http://etherx.jabber.org/streams' version='1.0'>", domain)
    _, err := io.WriteString(xc.Conn, streamHeader)
    return err
}

func (xc *XMPPConnection) CloseStream() error {
    _, err := io.WriteString(xc.Conn, "</stream:stream>")
    return err
}
