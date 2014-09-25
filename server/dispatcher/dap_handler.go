package dispatcher

import (
	"errors"
	"fmt"
	"io"
	"time"

	"getmelange.com/dap"

	"encoding/hex"

	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/server"

	"code.google.com/p/go-uuid/uuid"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

// New Name for DAP Handler
func (m *Server) GetMessages(since uint64, owner string, context bool) ([]*dap.ResponseMessage, error) {
	msg, err := m.GetIncomingMessagesSince(since, owner)
	if err != nil {
		return nil, err
	}

	out := make([]*dap.ResponseMessage, len(msg))
	for i, v := range msg {
		data, err := v.ToDispatch(owner)
		if err != nil {
			m.HandleError(createError("(GetMessages:DAP) Marshalling message", err))
			continue
		}

		out[i] = dap.CreateResponseMessage(data, m.Key.Address, identity.CreateAddressFromString(v.To))
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

func (m *Server) RetrieveDataForUser(name string, author *identity.Address, forAddr *identity.Address) (*message.EncryptedMessage, io.ReadCloser) {
	auth, err := aws.EnvAuth()
	if err != nil {
		return nil, nil
	}

	s3cred := s3.New(auth, aws.USEast)
	buck := s3cred.Bucket("dispatcher_uploads")

	msg, err := m.GetDataMessageNamed(name, author.String(), forAddr.String())
	if err != nil {
		m.HandleError(&server.ServerError{
			Location: "Getting Data Message",
			Error:    err,
		})
		return nil, nil
	}

	r, err := buck.GetReader(msg.Path)
	if err != nil {
		m.HandleError(&server.ServerError{
			Location: "Getting Data Reader",
			Error:    err,
		})
		return nil, nil
	}

	data, err := msg.ToDispatch(forAddr.String())
	if err != nil {
		m.HandleError(createError("(RetrieveMessage) Marshalling message", err))
		return nil, nil
	}

	return data, r
}

func (m *Server) PublishDataMessage(name string, to []string, author string, message *message.EncryptedMessage, length uint64, r dap.ReadVerifier) error {
	// Get AWS Authentication
	auth, err := aws.EnvAuth()
	if err != nil {
		return err
	}

	randomName := hex.EncodeToString(uuid.NewRandom())

	s3cred := s3.New(auth, aws.USEast)
	buck := s3cred.Bucket("dispatcher-uploads")

	filename := fmt.Sprintf("/uploads/%s", randomName)

	err = buck.PutReader(filename, r, int64(length), "application/octet-stream", s3.Private)
	if err != nil {
		return err
	}

	if !r.Verify() {
		// Error with this file. We should probably delete it from buck.
		err = buck.Del(filename)
		if err != nil {
			return err
		}

		return errors.New("Couldn't verify the hash of the uploaded data file.")
	}

	return m.SaveDataMessage(name, to, author, message, filename, int64(length))
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
	bytes, err := message.ToBytes()
	if err != nil {
		return err
	}

	msg.Data = bytes

	_, err = m.dbmap.Update(msg)
	return err
}
