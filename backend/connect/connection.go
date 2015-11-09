package connect

import (
	"fmt"

	adErrors "airdispat.ch/errors"
	"airdispat.ch/message"
	"airdispat.ch/server"
	"airdispat.ch/wire"
)

// getMailFromMessage will extract the AirDispatch Mail message from a
// reconstructed message while checking for erros.
func getMailFromMessage(data []byte, messageType string, h message.Header) (*message.Mail, error) {
	if err := checkMessageType(wire.MailCode, messageType, data, h); err != nil {
		return nil, err
	}

	return message.CreateMailFromBytes(data, h)
}

// checkMessageType will ensure that the actual type of the message is
// not an AirDispatch error and that it is of expected type.
func checkMessageType(expectedType, actualType string, data []byte, h message.Header) error {
	if actualType == wire.ErrorCode {
		return adErrors.CreateErrorFromBytes(data, h)
	}

	if actualType != expectedType {
		return fmt.Errorf("Expected Type %s, but received %s.",
			actualType, expectedType)
	}

	return nil
}

// getMessageListFromServer will download a list of messages from the
// user specified by the AirDispatch location `loc`.
func (c *Client) getMessageListFromServer(since uint64, loc *Location) ([]*message.Mail, error) {
	transferRequest := server.CreateTransferMessageList(
		since,            // The earliest message to be returned
		c.Origin.Address, // The user sending the message
		loc.Server,       // The server performing the transfer
		loc.Author,       // The author of the messages
	)

	conn, err := message.ConnectToServer(loc.Server.Location)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = message.SignAndSendToConnection(transferRequest, c.Origin, loc.Server, conn)
	if err != nil {
		return nil, err
	}

	list, err := message.ReadMessageFromConnection(conn)
	if err != nil {
		return nil, err
	}

	bytes, typ, h, err := list.Reconstruct(c.Origin, true)
	if err != nil {
		return nil, err
	}

	if err := checkMessageType(wire.MessageListCode, typ, bytes, h); err != nil {
		return nil, err
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

		bytes, typ, h, err := data.Reconstruct(c.Origin, false)
		if err != nil {
			return output, err
		}

		mail, err := getMailFromMessage(bytes, typ, h)
		if err != nil {
			return output, err
		}

		output = append(output, mail)
	}

	return output, nil
}
