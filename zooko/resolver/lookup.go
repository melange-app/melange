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

func (c *Client) LookupApp(name string) (*AppRegistration, bool, error) {
	app := new(AppRegistration)
	found, err := c.lookupJSON(name, app)
	if err != nil || !found {
		return nil, found, err
	}

	if app.Name == "" || app.Manifest == "" {
		return nil, found, ErrInvalidRegistration
	}

	return app, found, nil
}

// Lookup will return a registration associated with a zooko server.
func (c *Client) Lookup(name string) (*Registration, bool, error) {
	reg := new(Registration)
	found, err := c.lookupJSON(name, reg)
	if err != nil || !found {
		return nil, found, err
	}

	return reg, true, c.validateRegistration(reg)
}

func (c *Client) validateRegistration(reg *Registration) error {
	// All registration fields are required.
	if reg.Address == "" ||
		len(reg.EncryptionKey) == 0 ||
		reg.Location == "" {
		return ErrInvalidRegistration
	}

	return nil
}

func (c *Client) lookupJSON(name string, obj interface{}) (bool, error) {
	val, found, err := c.lookupString(name)
	if err != nil || !found {
		return found, err
	}

	if err := json.Unmarshal([]byte(val), obj); err != nil {
		fmt.Println("Received error unmarshalling", name, err)
		return true, ErrInvalidRegistration
	}

	return true, nil
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
