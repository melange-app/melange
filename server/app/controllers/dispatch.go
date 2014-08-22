package controllers

import (
	"errors"

	"getmelange.com/app/models"

	adErrors "airdispat.ch/errors"
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"
	"airdispat.ch/server"
	"airdispat.ch/wire"
)

func sendAlert(r routing.Router, msgName string, from *identity.Identity, to string, serverAlias string) error {
	addr, err := r.LookupAlias(to, routing.LookupTypeALERT)
	if err != nil {
		return err
	}

	msgDescription := server.CreateMessageDescription(msgName, serverAlias, from.Address, addr)
	err = message.SignAndSend(msgDescription, from, addr)
	if err != nil {
		return err
	}
	return nil
}

func getProfile(r routing.Router, from *identity.Identity, to string, alias string) (*message.Mail, error) {
	return downloadMessage(r, "profile", from, to, alias)
}

func getAddresses(r routing.Router, to *models.Address) (server *identity.Address, author *identity.Address, err error) {
	if to.Fingerprint == "" {
		author, err = r.LookupAlias(to.Alias, routing.LookupTypeMAIL)
		if err != nil {
			return
		}
	} else {
		author = identity.CreateAddressFromString(to.Fingerprint)
	}

	server, err = r.LookupAlias(to.Alias, routing.LookupTypeTX)
	return
}

func downloadMessageFromServer(msgName string, from *identity.Identity, author *identity.Address, srv *identity.Address) (*message.Mail, error) {
	txMsg := server.CreateTransferMessage(msgName, from.Address, srv, author)
	bytes, typ, h, err := message.SendMessageAndReceive(txMsg, from, srv)
	if err != nil {
		return nil, err
	}

	if typ == wire.ErrorCode {
		return nil, adErrors.CreateErrorFromBytes(bytes, h)
	}

	if typ != wire.MailCode {
		return nil, errors.New("Wrong message type, got " + typ)
	}

	return message.CreateMailFromBytes(bytes, h)
}

func downloadMessage(r routing.Router, msgName string, from *identity.Identity, to string, serverAlias string) (*message.Mail, error) {
	srv, err := r.LookupAlias(serverAlias, routing.LookupTypeTX)
	if err != nil {
		return nil, err
	}

	author := identity.CreateAddressFromString(to)

	return downloadMessageFromServer(msgName, from, author, srv)
}

func downloadPublicMail(r routing.Router, since uint64, from *identity.Identity, to *models.Address) ([]*message.Mail, error) {
	srv, author, err := getAddresses(r, to)
	if err != nil {
		return nil, err
	}

	txMsg := server.CreateTransferMessageList(since, from.Address, srv, author)

	conn, err := message.ConnectToServer(srv.Location)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = message.SignAndSendToConnection(txMsg, from, srv, conn)
	if err != nil {
		return nil, err
	}

	list, err := message.ReadMessageFromConnection(conn)
	if err != nil {
		return nil, err
	}

	bytes, typ, h, err := list.Reconstruct(from, true)
	if err != nil {
		return nil, err
	}

	if typ == wire.ErrorCode {
		return nil, adErrors.CreateErrorFromBytes(bytes, h)
	}

	if typ != wire.MessageListCode {
		return nil, errors.New("(a) Wrong message type, got " + typ)
	}

	msgList, err := server.CreateMessageListFromBytes(bytes, h)
	if err != nil {
		return nil, err
	}

	var output []*message.Mail
	for i := uint64(0); i < msgList.Length; i++ {
		data, err := message.ReadMessageFromConnection(conn)
		if err != nil {
			return output, err
		}

		bytes, typ, h, err := data.Reconstruct(from, false)
		if err != nil {
			return output, err
		}

		if typ == wire.ErrorCode {
			return nil, adErrors.CreateErrorFromBytes(bytes, h)
		}

		if typ != wire.MailCode {
			return output, errors.New("(b) Wrong message type, got " + typ)
		}

		mail, err := message.CreateMailFromBytes(bytes, h)
		if err != nil {
			return output, err
		}

		output = append(output, mail)
	}

	return output, nil
}
