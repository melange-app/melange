package models

import (
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"
	"airdispat.ch/server"
	"airdispat.ch/wire"
	"errors"
	"github.com/coopernurse/gorp"
)

type Message struct {
	MessageId int
	Name      string
	From      string
	To        []string
	Timestamp int
}

func (m *Message) ToDispatch(dbm *gorp.DbMap, to string) (*message.Mail, error) {
	legal := false
	for _, v := range m.To {
		if v == to {
			legal = true
			break
		}
	}
	if !legal {
		return nil, errors.New("Unable to find message for user.")
	}

	var components []*Component
	_, err := dbm.Select(&components, "select * from dispatch_components where messageid = $1", m.MessageId)
	if err != nil {
		return nil, err
	}

	fromAddr := identity.CreateAddressFromString(m.From)
	toAddr := identity.CreateAddressFromString(to)
	newMessage := message.CreateMail(fromAddr, toAddr)

	for _, v := range components {
		newMessage.Components.AddComponent(message.CreateComponent(v.Name, v.Data))
	}
	return newMessage, nil
}

func (m *Message) ToDescription(to string) (*server.MessageDescription, error) {
	// Location: "SERVER LOCATION"
	legal := false
	for _, v := range m.To {
		if v == to {
			legal = true
			break
		}
	}
	if !legal {
		return nil, errors.New("Unable to find user for that message.")
	}

	fromAddr := identity.CreateAddressFromString(m.From)
	toAddr := identity.CreateAddressFromString(to)
	return server.CreateMessageDescription(m.Name, "LOCATION", fromAddr, toAddr), nil
}

type Component struct {
	ComponentId int
	MessageId   int
	Name        string
	Data        []byte
}

type Alert struct {
	AlertId   int
	From      string
	To        string
	Timestamp int64
	Location  string
	Name      string
}

func CreateAlertFromDescription(desc *server.MessageDescription) *Alert {
	h := desc.Header()
	return &Alert{
		From:      h.From.String(),
		To:        h.To.String(),
		Timestamp: h.Timestamp,
		Location:  desc.Location,
		Name:      desc.Name,
	}
}

func (a *Alert) DownloadMessageFromAlert(db *gorp.DbMap, r routing.Router) (*message.Mail, error) {
	addr, err := r.Lookup(a.From)
	if err != nil {
		return nil, err
	}

	sender := IdentityFromFingerprint(a.To, db)
	if sender == nil {
		return nil, errors.New("adsf")
	}

	msgDescription := server.CreateTransferMessage("testMessage", sender.Address, addr)

	data, messageType, h, err := message.SendMessageAndReceive(msgDescription, sender, addr)

	if messageType != wire.MailCode {
		return nil, errors.New("Unexpected message type.")
	}

	return message.CreateMailFromBytes(data, h)
}

func DownloadPublicMessages(since uint64, addr *identity.Address, from *identity.Identity) ([]*Alert, error) {
	t := server.CreateTransferMessageList(since, from.Address, addr)

	data, messageType, h, err := message.SendMessageAndReceive(t, from, addr)

	if messageType != wire.MessageListCode {
		return nil, errors.New("Unexpected message type.")
	}

	ml, err := server.CreateMessageListFromBytes(data, h)
	if err != nil {
		return nil, err
	}

	out := make([]*Alert, len(ml.Content))
	for i, v := range ml.Content {
		out[i] = CreateAlertFromDescription(v)
	}

	return out, nil
}
