package dispatcher

import (
	"melange/dap"
	"time"

	"airdispat.ch/identity"
	"airdispat.ch/message"
)

// New Name for DAP Handler
func (m *Server) GetMessages(since uint64, owner string, context bool) ([]*dap.ResponseMessage, error) {
	msg, err := m.GetIncomingMessagesSince(since, owner)
	if err != nil {
		return nil, err
	}

	out := make([]*dap.ResponseMessage, len(msg))
	for i, v := range msg {
		out[i] = dap.CreateResponseMessage(v.ToDispatch(owner), identity.CreateAddressFromString(owner), identity.CreateAddressFromString(v.To))
	}

	return out, nil
}

func (m *Server) Unregister(user string, keys map[string][]byte) error {
	return nil
}

func (m *Server) Register(user string, keys map[string][]byte) error {
	obj := &User{
		Name:         string(keys["name"]),
		Receiving:    user,
		RegisteredOn: time.Now().Unix(),
	}

	err := m.dbmap.Insert(obj)
	if err != nil {
		return err
	}

	id := &Identity{
		Owner:   obj.Id,
		Signing: user,
	}
	return m.dbmap.Insert(id)
}

func (m *Server) PublishMessage(name string, to []string, author string, message *message.EncryptedMessage, alerted bool) error {
	messageType := TypeOutgoingPublic
	if alerted {
		messageType = TypeOutgoingPrivate
	}
	return m.SaveMessage(name, to, author, message, messageType)
}

func (m *Server) UpdateMessage(name string, author string, message *message.EncryptedMessage) error {
	msg, err := m.GetAnyMessageWithName(name, author)
	if err != nil {
		return err
	}

	// Load New Information
	msg.Data = message.Data
	msg.EncKey = message.EncryptionKey
	msg.EncType = message.EncryptionType

	_, err = m.dbmap.Update(msg)
	return err
}
