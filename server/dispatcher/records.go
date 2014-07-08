package dispatcher

import (
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"fmt"
	"strings"
	"time"
)

// Table Names
const (
	TableNameUser     = "melange_user"
	TableNameIdentity = "melange_identity"
	TableNameMessage  = "melange_message"
	TableNameStorage  = "melange_storage"
)

// Primary Key Names
const (
	PKUser     = "Id"
	PKIdentity = "Id"
	PKMessage  = "Id"
	PKStorage  = "Id"
)

// Create the Table Objects
func (s *Server) CreateTables() {
	// User Management
	s.dbmap.AddTableWithName(User{}, TableNameUser).SetKeys(true, PKUser)
	s.dbmap.AddTableWithName(Identity{}, TableNameIdentity).SetKeys(true, PKIdentity)

	// Message Management
	s.dbmap.AddTableWithName(Message{}, TableNameMessage).SetKeys(true, PKMessage)

	// Storage
	s.dbmap.AddTableWithName(Storage{}, TableNameStorage).SetKeys(true, PKStorage)
}

// Outgoing Messages
type Message struct {
	Id int
	// Recipient Information
	To     string
	Sender string
	// Message Information
	Name string
	Data []byte
	Type int
	// Encryption Information
	EncKey  []byte
	EncType []byte
	// Metadata
	Received int64
	// Transient
	allowed []string
}

const QueryOutgoingNamed = "select * from " + TableNameMessage + "o where o.Sender = :owner and o.Name = :name and o.To like :recv and o.Type = 1"
const QueryOutgoingPublic = "select * from " + TableNameMessage + "o where o.Sender = :owner and (o.To like :recv or o.To = '') and o.Received > :time and o.Type = 0"

// Return Outgoing Message Named
func (m *Server) GetOutgoingMessageWithName(name string, owner string, receiver string) (*Message, error) {
	var result *Message

	// Create the Query
	err := m.dbmap.SelectOne(&result, QueryOutgoingNamed,
		map[string]interface{}{
			"name":  name,
			"recv":  fmt.Sprintf("%%%s%%", receiver),
			"owner": owner,
		})
	if err != nil {
		return nil, err
	}

	return result, err
}

// Return Outgoing Public Messages for a Receiver
func (m *Server) GetOutgoingPublicMessagesFor(since uint64, owner string, receiver string) ([]*Message, error) {
	var results []*Message

	// Create the Query
	_, err := m.dbmap.Select(&results, QueryOutgoingPublic,
		map[string]interface{}{
			"recv":  fmt.Sprintf("%%%s%%", receiver),
			"owner": owner,
			"time":  since,
		})
	if err != nil {
		return nil, err
	}

	return results, err
}

const (
	TypeOutgoingPublic = iota
	TypeOutgoingPrivate
	TypeIncoming
)

// Save Outgoing Message
func (m *Server) SaveMessage(name string, to []string, from string, message *message.EncryptedMessage, messageType int) error {
	out := &Message{
		To:       strings.Join(to, ","),
		Sender:   from,
		Name:     name,
		Data:     message.Data,
		Type:     messageType,
		EncKey:   message.EncryptionKey,
		EncType:  message.EncryptionType,
		Received: time.Now().Unix(),
	}
	return m.dbmap.Insert(out)
}

// Change Outgoing Message into Encrypted Message
func (o *Message) ToDispatch(retriever string) *message.EncryptedMessage {
	return &message.EncryptedMessage{
		Data:           o.Data,
		EncryptionKey:  o.EncKey,
		EncryptionType: o.EncType,
		To:             identity.CreateAddressFromString(retriever),
	}
}

// Save Outgoing Message
func (m *Server) SaveIncomingMessage(message *message.EncryptedMessage) error {
	return m.SaveMessage("", []string{message.To.String()}, "", message, TypeIncoming)
}

const QueryIncoming = "select * from " + TableNameMessage + "o where o.To = :owner and o.Received > :time and o.Type = 2"

// Return Incoming Messages Since
func (m *Server) GetIncomingMessagesSince(since uint64, owner string) ([]*Message, error) {
	var results []*Message

	// Create the Query
	_, err := m.dbmap.Select(&results, QueryIncoming,
		map[string]interface{}{
			"owner": owner,
			"time":  since,
		})
	if err != nil {
		return nil, err
	}

	return results, err
}

type Storage struct {
	Id    int
	Key   string
	Value string
	Owner int
}

type User struct {
	Id           int
	Name         string
	Receiving    string
	RegisteredOn time.Time
}

type Identity struct {
	// Signing Key and Encryption Key
	Id         int
	Owner      int
	Signing    string
	Encrypting []byte
}
