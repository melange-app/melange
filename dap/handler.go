package dap

import (
	"bytes"
	"errors"
	"time"

	"hash"
	"io"
	"net"
	"strings"
	"sync"

	"getmelange.com/dap/wire"

	"crypto/sha256"

	"fmt"

	"airdispat.ch/identity"
	"airdispat.ch/message"
	"code.google.com/p/goprotobuf/proto"
)

const codePrefix = "DAP-"

type Delegate interface {
	// Account
	Register(addr string, keys map[string][]byte) error
	Unregister(addr string, keys map[string][]byte) error
	// Message
	GetMessages(since uint64, owner string, context bool) ([]*ResponseMessage, error)
	PublishMessage(name string, to []string, author string, message *message.EncryptedMessage, alerted bool) error
	UpdateMessage(name string, author string, message *message.EncryptedMessage) error
	// Sent Messages
	GetSentMessages(since uint64, owner string, context bool) ([]*ResponseMessage, error)
	// AD Data
	PublishDataMessage(name string, to []string, author string, message *message.EncryptedMessage, length uint64, r ReadVerifier) error
	// Data
	GetData(owner string, key string) ([]byte, error)
	SetData(owner string, key string, data []byte) error
	// Account Link Requests
}

type ReadVerifier interface {
	io.Reader
	Verify() bool
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
func (h *Handler) HandleMessage(typ string, data []byte, head message.Header, conn net.Conn) ([]message.Message, error) {
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
	case wire.PublishDataMessageCode:
		// Handle PublishDataMessage
		unmarsh := &wire.PublishDataMessage{}
		err := proto.Unmarshal(data, unmarsh)
		if err != nil {
			return nil, err
		}
		return h.PublishDataMessage(unmarsh, head, conn)
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

	// I don't trust servers to not delete the requested information
	// so we will just handle everything in memory (yay!).
	// Should be fine so long as millions of people don't try to link
	// at the same time... Haha.
	case wire.EnableLinkCode:
		return h.EnableLink(head)
	case wire.LinkRequestCode:
		unmarsh := &wire.LinkData{}
		err := proto.Unmarshal(data, unmarsh)
		if err != nil {
			return nil, err
		}

		return h.LinkRequest(unmarsh, head)
	case wire.LinkKeyCode:
		unmarsh := &wire.LinkData{}
		err := proto.Unmarshal(data, unmarsh)
		if err != nil {
			return nil, err
		}

		return h.LinkKey(unmarsh, head)
	case wire.LinkTransferCode:
		unmarsh := &wire.LinkTransfer{}
		err := proto.Unmarshal(data, unmarsh)
		if err != nil {
			return nil, err
		}

		return h.LinkTransfer(unmarsh, head)
	}
	return nil, fmt.Errorf("Cannot handle type (%s). That shouldn't happen.", typ)
}

var theLinkManager = &LinkManager{
	EnabledAccounts: make(map[string]time.Time),
	EnabledLock:     &sync.RWMutex{},

	Requests:     make(map[string][]byte),
	RequestsLock: &sync.RWMutex{},

	KeyTransfers:  make(map[string][]byte),
	TransfersLock: &sync.RWMutex{},
}

type LinkManager struct {
	// Track which accounts are accepting link
	// requests at the current moment.
	// map: requestee -> time enabled
	EnabledAccounts map[string]time.Time
	EnabledLock     *sync.RWMutex

	// Track the link requests
	// map: requestee -> request payload
	Requests     map[string][]byte
	RequestsLock *sync.RWMutex

	// Track the Transfers
	// map: requester -> key payload
	KeyTransfers  map[string][]byte
	TransfersLock *sync.RWMutex
}

func (h *Handler) EnableLink(head message.Header) ([]message.Message, error) {
	theLinkManager.EnabledLock.Lock()
	defer theLinkManager.EnabledLock.Unlock()

	theLinkManager.EnabledAccounts[head.From.String()] = time.Now().Add(5 * time.Minute)

	return CreateResponse(0, "OK", h.Key.Address, head.From), nil
}

func (h *Handler) LinkKey(msg *wire.LinkData, head message.Header) ([]message.Message, error) {
	theLinkManager.TransfersLock.Lock()
	defer theLinkManager.TransfersLock.Unlock()

	theLinkManager.KeyTransfers[*msg.For] = msg.Payload

	return CreateResponse(0, "OK", h.Key.Address, head.From), nil
}

func (h *Handler) LinkRequest(msg *wire.LinkData, head message.Header) ([]message.Message, error) {
	theLinkManager.RequestsLock.Lock()
	defer theLinkManager.RequestsLock.Unlock()

	theLinkManager.EnabledLock.RLock()
	defer theLinkManager.EnabledLock.RUnlock()
	if t, ok := theLinkManager.EnabledAccounts[*msg.For]; !ok || time.Now().After(t) {
		return nil, errors.New("Account not enabled for linking.")
	}

	theLinkManager.Requests[*msg.For] = msg.Payload

	return CreateResponse(0, "OK", h.Key.Address, head.From), nil
}

func (h *Handler) LinkTransfer(msg *wire.LinkTransfer, head message.Header) ([]message.Message, error) {
	if msg.Request != nil {
		theLinkManager.RequestsLock.RLock()
		defer theLinkManager.RequestsLock.RUnlock()

		r, ok := theLinkManager.Requests[head.From.String()]
		if !ok {
			return CreateResponse(
				12,
				"No request for that address",
				h.Key.Address,
				head.From,
			), nil
		}
		defer delete(theLinkManager.Requests, head.From.String())

		// Send the Response
		return CreateDataResponse(
			10,
			"Request Found",
			h.Key.Address,
			head.From,
			r,
		), nil
	} else if msg.Approved != nil {
		theLinkManager.TransfersLock.RLock()
		defer theLinkManager.TransfersLock.RUnlock()

		r, ok := theLinkManager.KeyTransfers[head.From.String()]
		if !ok {
			return CreateResponse(
				12,
				"No key transfer for that address",
				h.Key.Address,
				head.From,
			), nil
		}
		defer delete(theLinkManager.KeyTransfers, head.From.String())

		// Send the Response
		return CreateDataResponse(
			10,
			"Key Transfer Found",
			h.Key.Address,
			head.From,
			r,
		), nil
	}

	return CreateResponse(11, "No type of transfer specified.", h.Key.Address, head.From), nil
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
	return CreateResponse(0, "OK", h.Key.Address, head.From), nil
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
	return CreateResponse(0, "OK", h.Key.Address, head.From), nil
}

// Return all Messages received after `since` in sequence.
func (h *Handler) DownloadMessages(r *wire.DownloadMessages, head message.Header) ([]message.Message, error) {
	var (
		responses []*ResponseMessage
		err       error
	)

	// If the 'GetSent' flag is set, then we want to retrieve all sent messages.
	if r.GetSent() {
		responses, err = h.Delegate.GetSentMessages(r.GetSince(), head.From.String(), r.GetContext())
		if err != nil {
			return nil, err
		}
	} else {
		responses, err = h.Delegate.GetMessages(r.GetSince(), head.From.String(), r.GetContext())
		if err != nil {
			return nil, err
		}
	}

	out := make([]message.Message, len(responses))
	for i, v := range responses {
		out[i] = v
	}

	return CreateResponse(0, "OK", h.Key.Address, head.From, out...), nil
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

	return CreateResponse(0, "OK", h.Key.Address, head.From), nil
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

	return CreateResponse(0, "OK", h.Key.Address, head.From), nil
}

type dataReader struct {
	io.Reader

	expectedHash []byte
	runningHash  hash.Hash
}

func createDataReader(r io.Reader, length int64, hash []byte) *dataReader {
	newHash := sha256.New()
	return &dataReader{
		Reader:       io.TeeReader(io.LimitReader(r, length), newHash),
		expectedHash: hash,
		runningHash:  newHash,
	}
}

func (d *dataReader) Verify() bool {
	return bytes.Equal(d.runningHash.Sum(nil), d.expectedHash)
}

// Publish a message on Delegate.
func (h *Handler) PublishDataMessage(r *wire.PublishDataMessage, head message.Header, conn net.Conn) ([]message.Message, error) {
	msg, err := message.CreateEncryptedMessageFromBytes(r.GetHeader())
	if err != nil {
		return nil, err
	}

	// Name, To, Author, Message
	err = h.Delegate.PublishDataMessage(
		r.GetName(),
		r.GetTo(),
		head.From.String(),
		msg,
		r.GetLength(),
		createDataReader(conn, int64(r.GetLength()), r.GetHash()))
	if err != nil {
		return nil, err
	}

	return CreateResponse(0, "OK", h.Key.Address, head.From), nil
}

// Set Data
func (h *Handler) Data(r *wire.Data, head message.Header) ([]message.Message, error) {
	err := h.Delegate.SetData(head.From.String(), r.GetKey(), r.GetData())
	if err != nil {
		return nil, err
	}

	return CreateResponse(0, "OK", h.Key.Address, head.From), nil
}

// Get Data
func (h *Handler) GetData(r *wire.GetData, head message.Header) ([]message.Message, error) {
	data, err := h.Delegate.GetData(head.From.String(), r.GetKey())
	if err != nil {
		return nil, err
	}

	return CreateDataResponse(0, "OK", h.Key.Address, head.From, data), nil
}
