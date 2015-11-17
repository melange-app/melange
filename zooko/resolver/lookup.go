package resolver

import (
	"encoding/json"
	"errors"
	"fmt"

	adErrors "airdispat.ch/errors"
	adMessage "airdispat.ch/message"
	"airdispat.ch/wire"
	"code.google.com/p/goprotobuf/proto"
	"getmelange.com/zooko/config"
	"getmelange.com/zooko/message"
)

var (
	ErrInvalidRegistration = errors.New("zooko/resolver: returned value is not a valid zooko registration")
)

// Lookup will return a registration associated with a zooko server.
func (c *Client) Lookup(name string) (*Registration, bool, error) {
	val, found, err := c.lookupString(name)
	if err != nil || !found {
		return nil, found, err
	}

	reg := new(Registration)
	if err := json.Unmarshal([]byte(val), reg); err != nil {
		fmt.Println("Received error unmarshalling", name, err)
		return nil, true, ErrInvalidRegistration
	}

	// All registration fields are required.
	if reg.Address == "" ||
		len(reg.EncryptionKey) == 0 ||
		reg.Location == "" {
		return nil, true, ErrInvalidRegistration
	}

	return reg, true, err
}

// Lookup will check to see if the Zooko server
func (c *Client) lookupString(name string) (string, bool, error) {
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

func (c *Client) LookupPrefix(prefix string) ([]string, error) {
	trueVal := true
	msg, err := message.CreateMessage(&message.LookupName{
		Lookup: &prefix,
		Prefix: &trueVal,
	}, c.Origin, config.ServerAddress())
	if err != nil {
		return nil, err
	}

	data, typ, h, err := adMessage.SendMessageAndReceiveWithTimestamp(
		msg, c.Origin, config.ServerAddress())
	if err != nil {
		return nil, err
	} else if typ == wire.ErrorCode {
		return nil, adErrors.CreateErrorFromBytes(data, h)
	}

	if typ != message.TypeListNames {
		return nil, errors.New("zooko/resolver: received incorrect message reply")
	}

	parsed := new(message.ListName)
	err = proto.Unmarshal(data, parsed)
	if err != nil {
		return nil, err
	}

	return parsed.Name, nil
}
