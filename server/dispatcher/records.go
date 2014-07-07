package dispatcher

import (
	"airdispat.ch/message"
	"airdispat.ch/server"
	"github.com/coopernurse/gorp"
	"time"
)

func (s *Server) CreateTables() {
	return
}

type OutgoingMessage struct {
	Owner    int
	Name     string
	To       []string
	Data     []byte
	Received time.Time
	Alerted  bool
}

func (o *OutgoingMessage) ToDispatch(dbmap *gorp.DbMap, retriever string) (*message.Mail, error) {
	return nil, nil
}

func (o *OutgoingMessage) ToDescription(retriever string) (*server.MessageDescription, error) {
	return nil, nil
}

type IncomingAlert struct {
	Owner    int
	Data     []byte
	Received time.Time
}

func CreateAlertFromDescription(m *server.MessageDescription) *IncomingAlert {
	return nil
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
