package dispatcher

import (
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/server"
	"fmt"
	"strings"
	"time"
)

// Table Names
const (
	TableNameUser     = "melange_user"
	TableNameIdentity = "melange_identity"
	TableNameOutgoing = "melange_outgoing"
	TableNameIncoming = "melange_incoming"
	TableNameStorage  = "melange_storage"
)

// Primary Key Names
const (
	PKUser     = "UserId"
	PKIdentity = "IdentityId"
	PKOutgoing = "MessageId"
	PKIncoming = "AlertId"
	PKStorage  = "RecordId"
)

// Create the Table Objects
func (s *Server) CreateTables() {
	// User Management
	s.dbmap.AddTableWithName(User{}, TableNameUser).SetKeys(true, PKUser)
	s.dbmap.AddTableWithName(Identity{}, TableNameIdentity).SetKeys(true, PKIdentity)

	// Message Management
	s.dbmap.AddTableWithName(OutgoingMessage{}, TableNameOutgoing).SetKeys(true, PKOutgoing)
	s.dbmap.AddTableWithName(IncomingAlert{}, TableNameIncoming).SetKeys(true, PKIncoming)

	// Storage
	s.dbmap.AddTableWithName(Storage{}, TableNameStorage).SetKeys(true, PKStorage)
}

// Outgoing Messages
type OutgoingMessage struct {
	Id int
	// Recipient Information
	To     string
	Sender string
	// Message Information
	Name    string
	Data    []byte
	Alerted int
	// Encryption Information
	EncKey  []byte
	EncType []byte
	// Metadata
	Received int64
	// Transient
	allowed []string
}

const QueryOutgoingNamed = "select * from " + TableNameOutgoing + "o where o.Sender = :owner and o.Name = :name and o.To like :recv"
const QueryOutgoingPublic = "select * from " + TableNameOutgoing + "o where o.Sender = :owner and (o.To like :recv or o.To = '') and o.Received > :time and o.Alerted = 0"

// Return Outgoing Message Named
func (m *Server) GetOutgoingMessageWithName(name string, owner string, receiver string) (*OutgoingMessage, error) {
	var result *OutgoingMessage

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
func (m *Server) GetOutgoingPublicMessagesFor(since uint64, owner string, receiver string) ([]*OutgoingMessage, error) {
	var results []*OutgoingMessage

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

// Save Outgoing Message
func (m *Server) SaveOutgoingMessage(name string, to []string, from string, message *message.EncryptedMessage, alerted bool) error {
	// Convert Bool to Int
	sqlAlerted := 0
	if alerted {
		sqlAlerted = 1
	}

	out := &OutgoingMessage{
		To:       strings.Join(to, ","),
		Sender:   from,
		Name:     name,
		Data:     message.Data,
		Alerted:  sqlAlerted,
		EncKey:   message.EncryptionKey,
		EncType:  message.EncryptionType,
		Received: time.Now().Unix(),
	}
	return m.dbmap.Insert(out)
}

// Change Outgoing Message into Encrypted Message
func (o *OutgoingMessage) ToDispatch(retriever string) *message.EncryptedMessage {
	return &message.EncryptedMessage{
		Data:           o.Data,
		EncryptionKey:  o.EncKey,
		EncryptionType: o.EncType,
		To:             identity.CreateAddressFromString(retriever),
	}
}

type IncomingAlert struct {
	Id       int
	Owner    int
	Data     []byte
	Received time.Time
}

func CreateAlertFromDescription(m *server.MessageDescription) *IncomingAlert {
	return nil
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
