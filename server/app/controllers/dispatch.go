package controllers

import (
	"errors"

	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"
	"airdispat.ch/server"
	"airdispat.ch/wire"
)

func sendAlert(r routing.Router, msgName string, from *identity.Identity, to string, location string) error {
	addr, err := r.LookupAlias(to, routing.LookupTypeALERT)
	if err != nil {
		return err
	}

	msgDescription := server.CreateMessageDescription(msgName, location, from.Address, addr)
	err = message.SignAndSend(msgDescription, from, addr)
	if err != nil {
		return err
	}
	return nil
}

func getProfile(r routing.Router, from *identity.Identity, to string) (*message.Mail, error) {
	return downloadMessage(r, "profile", from, to, "")
}

func downloadMessage(r routing.Router, msgName string, from *identity.Identity, to string, toServer string) (*message.Mail, error) {
	addr, err := r.LookupAlias(to, routing.LookupTypeTX)
	if err != nil {
		return nil, err
	}
	if toServer != "" {
		addr.Location = toServer
	}

	txMsg := server.CreateTransferMessage(msgName, from.Address, addr)
	bytes, typ, h, err := message.SendMessageAndReceiveWithoutTimestamp(txMsg, from, addr)
	if err != nil {
		return nil, err
	}

	if typ != wire.MailCode {
		return nil, errors.New("Wrong message type.")
	}

	return message.CreateMailFromBytes(bytes, h)
}

func downloadPublicMail(r routing.Router, since uint64, from *identity.Identity, to string) ([]*message.Mail, error) {
	addr, err := r.LookupAlias(to, routing.LookupTypeTX)
	if err != nil {
		return nil, err
	}

	txMsg := server.CreateTransferMessageList(since, from.Address, addr)

	conn, err := message.ConnectToServer(addr.Location)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = message.SignAndSendToConnection(txMsg, from, addr, conn)
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

	if typ != wire.MessageListCode {
		return nil, errors.New("Wrong message type.")
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

		if typ != wire.MailCode {
			return output, errors.New("Wrong message type.")
		}

		mail, err := message.CreateMailFromBytes(bytes, h)
		if err != nil {
			return output, err
		}

		output = append(output, mail)
	}

	return output, nil
}
