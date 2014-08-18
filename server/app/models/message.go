package models

import "github.com/huntaub/go-db"

// A full AirDispatch Message
type Message struct {
	Id db.PrimaryKey
	// Name registered on Server
	Name string
	// Cleared Addresses
	To string
	// Fingerprint of Sender
	From string
	// Message Metadata
	Alert    bool
	Incoming bool
	Date     int64

	Components *db.HasMany `table:"component" on:"message"`
}

// CreateMessage(name, to, from, alert, component...)
// m.UpdateMessage()    [ Client -> Server ]
// m.RetrieveMessage()  [ Client <- Server ]
// m.Delete()

// An AirDispatch Component
type Component struct {
	Id      db.PrimaryKey
	Message *db.HasOne `table:"message"`
	Name    string
	Data    []byte
}
