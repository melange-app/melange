package connect

import (
	"fmt"

	mIdentity "getmelange.com/backend/models/identity"

	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"
	"airdispat.ch/server"
	"getmelange.com/dap"
)

const (
	errNameUnexpected = "Name enumerated in the message (%s) is different than retrieved (%s)."
)

// Location represents an AirDispatch location that a message can be
// sent to.
type Location struct {
	Author *identity.Address
	Server *identity.Address

	Model *mIdentity.Address
}

// Client is an object that allows for making arbitrary AirDispatch
// requests.
type Client struct {
	Router routing.Router
	Origin *identity.Identity

	*dap.Client
}

// SendAlert will deliver an AirDispatch alert message to a recipient.
//
// msgName - The name of the message that the recipient will be
//           alerted about. It should already be uploaded to the
//           Author's dispatcher.
//      to - The alias of the user who will receive the alert.
//  server - The alias of the server that will store the alert.
func (c *Client) SendAlert(msgName string, to string) error {
	addr, err := c.Router.LookupAlias(to, routing.LookupTypeALERT)
	if err != nil {
		return err
	}

	msgDescription := server.CreateMessageDescription(msgName, c.Client.Server.Alias, c.Origin.Address, addr)
	return message.SignAndSend(msgDescription, c.Origin, addr)
}

// GetProfile will return the profile message of the requested user.
//
//    to - The alias of the profile Author.
// alias - The server alias of the profile Author.
func (c *Client) GetProfile(to string, alias string) (*message.Mail, error) {
	return c.GetMessage("profile", to, alias)
}

// GetLocation will determine the AirDispatch location of a User based
// on the stored address information.
func (c *Client) GetLocation(to *mIdentity.Address) (*Location, error) {
	var (
		author *identity.Address
		err    error
	)

	if to.Fingerprint == "" {
		author, err = c.Router.LookupAlias(to.Alias, routing.LookupTypeMAIL)
		if err != nil {
			return nil, err
		}
	} else {
		author = identity.CreateAddressFromString(to.Fingerprint)
	}

	server, err := c.Router.LookupAlias(to.Alias, routing.LookupTypeTX)

	return &Location{
		Author: author,
		Server: server,
	}, err
}

// getMessageFromServer will return the message named `msgName` from
// the AirDispatch location specified in `loc`.
func (c *Client) getMessageFromServer(msgName string, loc *Location) (*message.Mail, error) {
	txMsg := server.CreateTransferMessage(msgName, c.Origin.Address, loc.Server, loc.Author)
	bytes, typ, h, err := message.SendMessageAndReceive(txMsg, c.Origin, loc.Server)
	if err != nil {
		return nil, err
	}

	msg, err := getMailFromMessage(bytes, typ, h)
	if err != nil {
		return nil, err
	}

	if msg.Name != "" && msg.Name != msgName {
		return nil,
			fmt.Errorf(errNameUnexpected, msg.Name, msgName)
	}

	// TODO: Check for Name Accoutability Here
	return msg, err
}

// GetMessage will return the messsage named `msgName` by the author
// with alias `to` utilizing a dispatcher with alias `serverAlias`.
func (c *Client) GetMessage(msgName string, to string, serverAlias string) (*message.Mail, error) {
	toLookup := serverAlias
	if toLookup == "" {
		toLookup = to
	}

	srv, err := c.Router.LookupAlias(serverAlias, routing.LookupTypeTX)
	if err != nil {
		return nil, err
	}

	author := identity.CreateAddressFromString(to)

	return c.getMessageFromServer(msgName, &Location{
		Author: author,
		Server: srv,
	})
}

// GetPublicMessages will return the public messages that have been
// authored by `to` since the Unix timestamp specified by `since`.
func (c *Client) GetPublicMessages(since uint64, to *mIdentity.Address) ([]*message.Mail, error) {
	loc, err := c.GetLocation(to)
	if err != nil {
		return nil, err
	}

	return c.getMessageListFromServer(since, loc)
}
