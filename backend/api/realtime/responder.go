package realtime

import (
	"encoding/json"

	"getmelange.com/backend/info"
)

// Request objects for realtime will give access to a channel to
// respond to messages.
type Request struct {
	*Message

	Response    chan *Message
	Environment *info.Environment
}

// Responder is an object that can respond to a Realtime request. It
// should return nil if the object cannot be satisfied.
type Responder interface {
	Handle(*Request) bool
}

// Responders represents a list of responders that can be tried in
// order.
type Responders []Responder

// Handle implements the Responder interface on the list.
func (r Responders) Handle(req *Request) bool {
	for _, v := range r {
		if v.Handle(req) {
			return true
		}
	}

	return false
}

// Message represents a JSON message passed across the wire.
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func mustCreateMessage(typ string, obj interface{}) *Message {
	msg, err := CreateMessage(typ, obj)
	if err != nil {
		panic(err)
	}

	return msg
}

func CreateMessage(typ string, obj interface{}) (*Message, error) {
	var err error

	m := &Message{
		Type: typ,
	}

	m.Data, err = json.Marshal(obj)

	return m, err
}

func (m *Message) Unmarshal(d interface{}) error {
	return json.Unmarshal(m.Data, d)
}
