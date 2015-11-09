package connect

import (
	"fmt"
	"net"

	"airdispat.ch/message"
	"airdispat.ch/server"
	"airdispat.ch/wire"

	mIdentity "getmelange.com/backend/models/identity"
)

func (c *Client) GetDataMessage(msgName string, to string) (*message.DataMessage, net.Conn, error) {
	loc, err := c.GetLocation(&mIdentity.Address{
		Alias: to,
	})
	if err != nil {
		return nil, nil, err
	}

	conn, err := message.ConnectToServer(loc.Server.Location)
	if err != nil {
		return nil, nil, err
	}

	txMsg := server.CreateTransferMessage(msgName, c.Origin.Address, loc.Server, loc.Author)
	txMsg.Data = true

	err = message.SignAndSendToConnection(txMsg, c.Origin, loc.Server, conn)
	if err != nil {
		return nil, nil, err
	}

	enc, err := message.ReadMessageFromConnection(conn)
	if err != nil {
		return nil, nil, err
	}

	by, typ, h, err := enc.Reconstruct(c.Origin, false)
	if err != nil {
		return nil, nil, err
	}

	if err := checkMessageType(wire.DataCode, typ, by, h); err != nil {
		return nil, nil, err
	}

	dataMessage, err := message.CreateDataMessageFromBytes(by, h)
	if err != nil {
		return nil, nil, err
	}

	if dataMessage.Name != msgName {
		return nil, nil, fmt.Errorf("Name of message (%s) is different than expected (%s).",
			dataMessage.Name, msgName)
	}

	return dataMessage, conn, nil
}
