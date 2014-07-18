package models

import (
	"time"
)

// A full AirDispatch Message
type Message struct {
	// Name registered on Server
	Name string
	// Cleared Addresses
	To []string
	// Fingerprint of Sender
	From string
	// Message Metadata
	Alert    bool
	Incoming bool
	Date     *time.Time
}

// CreateMessage(name, to, from, alert, component...)
// m.UpdateMessage()    [ Client -> Server ]
// m.RetrieveMessage()  [ Client <- Server ]
// m.Delete()

// An AirDispatch Component
type Component struct {
	MessageId int
	Name      string
	Data      []byte
}
