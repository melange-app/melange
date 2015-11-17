package resolver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"code.google.com/p/goprotobuf/proto"

	adErrors "airdispat.ch/errors"
	adMessage "airdispat.ch/message"
	"airdispat.ch/wire"

	"getmelange.com/zooko/account"
	"getmelange.com/zooko/config"
	"getmelange.com/zooko/message"
)

func (c *Client) getRegistrationResponse(msg adMessage.Message) error {
	data, typ, h, err := adMessage.SendMessageAndReceiveWithTimestamp(msg, c.Origin, config.ServerAddress())
	if err != nil {
		return err
	} else if typ == wire.ErrorCode {
		return adErrors.CreateErrorFromBytes(data, h)
	} else if typ != message.TypeRegistrationResponse {
		return errors.New("zooko: received message with incorrect type")
	}

	response := new(message.RegistrationResponse)
	if err := proto.Unmarshal(data, response); err != nil {
		return err
	}

	if !*response.Success {
		return fmt.Errorf("zooko/resolver: got error registering: %s", *response.Information)
	}

	return nil
}

// Register will handle the registration of a new name in the Zooko server.
func (c *Client) Register(name string, reg interface{}, acc *account.Account) error {
	if _, found, _ := c.lookupString(name); found {
		return errors.New("zooko: cannot register name that already exists")
	}

	jsonRegistration, err := json.Marshal(reg)
	if err != nil {
		return err
	}

	if err := c.checkAccountBalance(acc, true); err != nil {
		return err
	}

	nameNew, rand, err := acc.CreateNameNew(name)
	if err != nil {
		return err
	}

	// Temporarily commit to the new transaction so that we can
	// successfully create the name_firstupdate transaction.
	oldUnspent := acc.Unspent
	acc.Commit(nameNew)
	undoCommit := func() {
		acc.Unspent = oldUnspent
	}

	nameUpdate, err := acc.CreateNameFirstUpdate(rand, name, string(jsonRegistration))
	if err != nil {
		undoCommit()
		return err
	}

	nameNewBytes := &bytes.Buffer{}
	if err := nameNew.MsgTx.Serialize(nameNewBytes); err != nil {
		undoCommit()
		return err
	}

	nameFirstUpdateBytes := &bytes.Buffer{}
	if err := nameUpdate.MsgTx.Serialize(nameFirstUpdateBytes); err != nil {
		undoCommit()
		return err
	}

	registrationMsg, err := message.CreateMessage(&message.RegisterName{
		Name:            &name,
		Value:           jsonRegistration,
		NameNew:         nameNewBytes.Bytes(),
		NameFirstupdate: nameFirstUpdateBytes.Bytes(),
	}, c.Origin, config.ServerAddress())
	if err != nil {
		undoCommit()
		return err
	}

	if err := c.getRegistrationResponse(registrationMsg); err != nil {
		undoCommit()
		return err
	}

	// Add the name_firstupdate transaction to the list of pending
	// transactions.
	acc.Pending = append(acc.Pending, nameUpdate)

	return nil
}

// Renew will handle renewing (or updating) a name that is already
// owned by acc on the Zooko server.
func (c *Client) Renew(name string, reg *Registration, acc *account.Account) error {
	jsonRegistration, err := json.Marshal(reg)
	if err != nil {
		return err
	}

	if err := c.checkAccountBalance(acc, false); err != nil {
		return err
	}

	nameUpdate, err := acc.CreateNameUpdate(name, string(jsonRegistration))
	if err != nil {
		return err
	}

	nameUpdateBytes := &bytes.Buffer{}
	if err := nameUpdate.Serialize(nameUpdateBytes); err != nil {
		return err
	}

	updateMsg, err := message.CreateMessage(&message.RenewName{
		Name:       &name,
		Value:      jsonRegistration,
		NameUpdate: nameUpdateBytes.Bytes(),
	}, c.Origin, config.ServerAddress())
	if err != nil {
		return err
	}

	if err := c.getRegistrationResponse(updateMsg); err != nil {
		return err
	}

	// Commit to the transaction since the server accepted it.
	acc.Commit(nameUpdate)

	return nil
}
