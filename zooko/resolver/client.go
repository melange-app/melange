package resolver

import (
	"errors"

	"code.google.com/p/goprotobuf/proto"

	adErrors "airdispat.ch/errors"
	"airdispat.ch/identity"
	adMessage "airdispat.ch/message"
	"airdispat.ch/wire"
	"getmelange.com/zooko/config"
	"getmelange.com/zooko/message"
)

type Client struct {
	Origin *identity.Identity
}

func CreateClient(id *identity.Identity) *Client {
	if id == nil {
		id, _ = identity.CreateIdentity()
	}

	return &Client{
		Origin: id,
	}
}

// Lookup will check to see if the Zooko server
func (c *Client) Lookup(name string) (string, bool, error) {
	msg, err := message.CreateMessage(&message.LookupName{
		Lookup: &name,
	}, c.Origin, config.ServerAddress())
	if err != nil {
		return "", false, err
	}

	data, typ, h, err := adMessage.SendMessageAndReceiveWithTimestamp(
		msg, c.Origin, config.ServerAddress())
	if err != nil {
		return "", false, err
	} else if typ == wire.ErrorCode {
		return "", false, adErrors.CreateErrorFromBytes(data, h)
	}

	if typ != message.TypeResolvedName {
		return "", false, errors.New("zooko/resolver: received incorrect message reply")
	}

	parsed := new(message.ResolvedName)
	err = proto.Unmarshal(data, parsed)
	if err != nil {
		return "", false, err
	}

	return string(parsed.Value), *parsed.Found, nil
}
