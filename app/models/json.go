package models

import (
	"time"
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"

	"getmelange.com/router"
)

// JSON Encoding
type JSONMessageList []*JSONMessage

func (m JSONMessageList) Len() int               { return len(m) }
func (m JSONMessageList) Less(i int, j int) bool { return m[i].Date.After(m[j].Date) }
func (m JSONMessageList) Swap(i int, j int)      { m[i], m[j] = m[j], m[i] }

type JSONMessage struct {
	Name       string                    `json:"name"`
	Date       time.Time                 `json:"date"`
	From       *JSONProfile              `json:"from"`
	To         []*JSONProfile            `json:"to"`
	Public     bool                      `json:"public"`
	Self       bool                      `json:"self"`
	Components map[string]*JSONComponent `json:"components"`
	Context    map[string]string         `json:"context"`
}

type JSONComponent struct {
	Binary []byte `json:"binary"`
	String string `json:"string"`
}

type JSONProfile struct {
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Alias       string `json:"alias"`
	Fingerprint string `json:"fingerprint"`
}

func (m *JSONMessage) ToModel(from *identity.Identity) (*Message, []*Component) {
	toAddrs := ""
	for i, v := range m.To {
		if i != 0 {
			toAddrs += ","
		}
		toAddrs += v.Alias
	}

	message := &Message{
		Name: m.Name,
		// Address
		To:   toAddrs,
		From: from.Address.String(),
		// Meta
		Date:     m.Date.Unix(),
		Incoming: false,
		Alert:    !m.Public,
	}

	components := make([]*Component, len(m.Components))
	i := 0
	for key, v := range m.Components {
		components[i] = &Component{
			Name: key,
		}

		if len(v.Binary) == 0 {
			components[i].Data = []byte(v.String)
		} else {
			components[i].Data = v.Binary
		}

		i++
	}

	return message, components
}

func (m *JSONMessage) ToDispatch(from *identity.Identity) (*message.Mail, []*identity.Address, error) {
	r := router.Router{
		Origin: from,
	}

	addrs := make([]*identity.Address, len(m.To))
	for i, v := range m.To {
		var err error
		addrs[i], err = r.LookupAlias(v.Alias, routing.LookupTypeMAIL)
		if err != nil {
			return nil, nil, err
		}
	}

	mail := message.CreateMail(from.Address, m.Date, m.Name, addrs...)

	for key, v := range m.Components {
		mail.Components.AddComponent(message.Component{
			Name: key,
			Data: []byte(v.String),
		})
	}

	return mail, addrs, nil
}
