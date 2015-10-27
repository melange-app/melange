package messages

import (
	"time"

	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"

	"getmelange.com/backend/connect"
)

// JSON Encoding
type JSONMessageList []*JSONMessage

func (m JSONMessageList) Len() int               { return len(m) }
func (m JSONMessageList) Less(i int, j int) bool { return m[i].Date.After(m[j].Date) }
func (m JSONMessageList) Swap(i int, j int)      { m[i], m[j] = m[j], m[i] }

// JSONMessage represents an AirDispatch message that is ready to be
// serialized or deserialized for the Melange client.
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

// JSONComponent is an AirDispatch component that is ready for the
// Melange client.
type JSONComponent struct {
	Binary []byte `json:"binary"`
	String string `json:"string"`
}

// JSONProfile is a Melange profile that is ready to be sent to the
// client as part of a message.
type JSONProfile struct {
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Alias       string `json:"alias"`
	Fingerprint string `json:"fingerprint"`
}

// ToModel will convert a JSONMessage into a model ready to be stored
// in the database.
func (m *JSONMessage) ToModel(from *connect.Client) (*Message, []*Component) {
	// Convert JSONProfile to a list of aliases.
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
		From: from.Origin.Address.String(),
		// Meta
		Date:     m.Date.Unix(),
		Incoming: false,
		Alert:    !m.Public,
	}

	// Convert the JSON components back to Model components.
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

// ToDispatch converts a JSONMessage into an AirDispatch message.
func (m *JSONMessage) ToDispatch(from *connect.Client) (*message.Mail, []*identity.Address, error) {
	// Convert the "To Field" into AirDispatch addresses.
	addrs := make([]*identity.Address, len(m.To))
	for i, v := range m.To {
		var err error
		addrs[i], err = from.Router.LookupAlias(v.Alias, routing.LookupTypeMAIL)
		if err != nil {
			return nil, nil, err
		}
	}

	// Create the AirDispatch mail message.
	mail := message.CreateMail(from.Origin.Address, m.Date, m.Name, addrs...)

	// Load the JSON components into the AirDispatch message.
	for key, v := range m.Components {
		mail.Components.AddComponent(message.Component{
			Name: key,
			Data: []byte(v.String),
		})
	}

	return mail, addrs, nil
}
