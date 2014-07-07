package dispatcher

import (
	"github.com/coopernurse/gorp"
	"time"
)

func (s *Server) CreateTables() {
	return
}

type OutgoingMessage struct {
	Owner    int
	To       []string
	Data     []byte
	Received time.Time
}

type IncomingAlert struct {
	Owner    int
	Data     []byte
	Received time.Time
}

type Storage struct {
	Key   string
	Value string
	Owner int
}

type User struct {
	Name         string
	Receiving    string
	RegisteredOn time.Time
}

type Identity struct {
	// Signing Key and Encryption Key
	Owner      int
	Signing    string
	Encrypting []byte
}
