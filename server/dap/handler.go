package dap

import (
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"code.google.com/p/goprotobuf/proto"
	"errors"
	"melange/dap/wire"
	"strings"
)

const codePrefix = "DAP-"

type Delegate interface {
	// Account
	Register(addr string, keys map[string][]byte) error
	Unregister(addr string, keys map[string][]byte) error
	// Message
	GetMessages(since uint64, owner string) ([]*message.EncryptedMessage, error)
	PublishMessage(name string, to []string, author string, message *message.EncryptedMessage, alerted bool) error
	UpdateMessage(name string, author string, message *message.EncryptedMessage) error
	// Data
	GetData(owner string, key string) ([]byte, error)
	SetData(owner string, key string, data []byte) error
}

type Handler struct {
	Key      *identity.Identity
	Delegate Delegate
}

// Handle Type
func (h *Handler) HandlesType(typ string) bool {
	return strings.HasPrefix(typ, codePrefix)
}

// Handle the DAP Request
func (h *Handler) HandleType(typ string, data []byte, head message.Header) ([]message.Message, error) {
	// I love that Golang doesn't have Generics. I promise!
	switch typ {
	case wire.RegisterCode:
		// Handle Registration
		unmarsh := &wire.Register{}
		err := proto.Unmarshal(data, unmarsh)
		if err != nil {
			return nil, err
		}
		return h.Register(unmarsh, head)
	case wire.UnregisterCode:
		// Handle Unregistration
		unmarsh := &wire.Unregister{}
		err := proto.Unmarshal(data, unmarsh)
		if err != nil {
			return nil, err
		}
		return h.Unregister(unmarsh, head)
	case wire.DownloadMessagesCode:
		// Handle DownloadMessages
		unmarsh := &wire.DownloadMessages{}
		err := proto.Unmarshal(data, unmarsh)
		if err != nil {
			return nil, err
		}
		return h.DownloadMessages(unmarsh, head)
	case wire.PublishMessageCode:
		// Handle PublishMessage
		unmarsh := &wire.PublishMessage{}
		err := proto.Unmarshal(data, unmarsh)
		if err != nil {
			return nil, err
		}
		return h.PublishMessage(unmarsh, head)
	case wire.UpdateMessageCode:
		// Handle UpdateMessage
		unmarsh := &wire.UpdateMessage{}
		err := proto.Unmarshal(data, unmarsh)
		if err != nil {
			return nil, err
		}
		return h.UpdateMessage(unmarsh, head)
	case wire.DataCode:
		// Handle Data
		unmarsh := &wire.Data{}
		err := proto.Unmarshal(data, unmarsh)
		if err != nil {
			return nil, err
		}
		return h.Data(unmarsh, head)
	case wire.GetDataCode:
		// Handle GetData
		unmarsh := &wire.GetData{}
		err := proto.Unmarshal(data, unmarsh)
		if err != nil {
			return nil, err
		}
		return h.GetData(unmarsh, head)
	}
	return nil, errors.New("Cannot handle type. That shouldn't happen.")
}

// Register a User on the Delegate
func (h *Handler) Register(r *wire.Register, head message.Header) ([]message.Message, error) {
	data := make(map[string][]byte)
	for _, v := range r.GetKeys() {
		data[v.GetKey()] = v.GetData()
	}
	err := h.Delegate.Register(head.From.String(), data)
	if err != nil {
		return nil, err
	}
	return CreateResponse(0, "OK", head.To, head.From), nil
}

// Unregister a User on the Delegate
func (h *Handler) Unregister(r *wire.Unregister, head message.Header) ([]message.Message, error) {
	data := make(map[string][]byte)
	for _, v := range r.GetKeys() {
		data[v.GetKey()] = v.GetData()
	}
	err := h.Delegate.Unregister(head.From.String(), data)
	if err != nil {
		return nil, err
	}
	return CreateResponse(0, "OK", head.To, head.From), nil
}

// Return all Messages received after `since` in sequence.
func (h *Handler) DownloadMessages(r *wire.DownloadMessages, head message.Header) ([]message.Message, error) {
	responses, err := h.Delegate.GetMessages(r.GetSince(), head.From.String())
	if err != nil {
		return nil, err
	}

	out := make([]message.Message, len(responses))
	for i, v := range responses {
		out[i] = CreateResponseMessage(v, head.To, head.From)
	}
	return CreateResponse(0, "OK", head.To, head.From, out...), nil
}

// Publish a message on Delegate.
func (h *Handler) PublishMessage(r *wire.PublishMessage, head message.Header) ([]message.Message, error) {
	msg, err := message.CreateEncryptedMessageFromBytes(r.GetData())
	if err != nil {
		return nil, err
	}

	// Name, To, Author, Message
	err = h.Delegate.PublishMessage(r.GetName(), r.GetTo(), head.From.String(), msg, r.GetAlert())
	if err != nil {
		return nil, err
	}

	return CreateResponse(0, "OK", head.To, head.From), nil
}

// Update a message on Delegate.
func (h *Handler) UpdateMessage(r *wire.UpdateMessage, head message.Header) ([]message.Message, error) {
	msg, err := message.CreateEncryptedMessageFromBytes(r.GetData())
	if err != nil {
		return nil, err
	}

	// Name, To, Author, Message
	err = h.Delegate.UpdateMessage(r.GetName(), head.From.String(), msg)
	if err != nil {
		return nil, err
	}

	return CreateResponse(0, "OK", head.To, head.From), nil
}

// Set Data
func (h *Handler) Data(r *wire.Data, head message.Header) ([]message.Message, error) {
	err := h.Delegate.SetData(head.From.String(), r.GetKey(), r.GetData())
	if err != nil {
		return nil, err
	}

	return CreateResponse(0, "OK", head.To, head.From), nil
}

// Get Data
func (h *Handler) GetData(r *wire.GetData, head message.Header) ([]message.Message, error) {
	data, err := h.Delegate.GetData(head.From.String(), r.GetKey())
	if err != nil {
		return nil, err
	}

	return CreateDataResponse(0, "OK", head.To, head.From, data), nil
}
