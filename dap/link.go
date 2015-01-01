package dap

import (
	"bytes"
	"errors"
	"math/big"
	"strconv"
	"time"

	"crypto/rand"

	"airdispat.ch/crypto"
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"code.google.com/p/goprotobuf/proto"
	"getmelange.com/dap/wire"
)

type LinkClient struct {
	*Client

	User *identity.Address
}

func CreateLinkClient(linkFrom *identity.Address, server *identity.Address) (*LinkClient, error) {
	id, err := identity.CreateIdentity()
	if err != nil {
		return nil, err
	}

	return &LinkClient{
		Client: CreateClient(id, server),
		User:   linkFrom,
	}, nil
}

type PendingLinkRequest struct {
	From         *identity.Address
	Verification string
}

// EnableLink tells the dispatcher that the account
// is ready to accept link requests.
func (c *Client) EnableLink() error {
	msg := &RawMessage{
		Message: &wire.EnableLink{},
		Code:    wire.EnableLinkCode,
		Head:    c.createHeader(c.Server),
	}

	_, err := c.sendAndGetResponse(msg)
	return err
}

// WaitForLinkRequeset will loop, querying the dispatcher
// every 10 seconds for a LinkRequest. It will error out after 5 minutes.
func (c *Client) WaitForLinkRequest() (*PendingLinkRequest, error) {
	queryTicker := time.NewTicker(10 * time.Second)
	stopChannel := time.After(5 * time.Minute)

	for {
		select {
		case <-stopChannel:
			queryTicker.Stop()
			break
		case <-queryTicker.C:
			r, err := c.GetLinkRequest()

			// Only return on a nil-error
			if err == nil {
				return r, err
			}
		}
	}

	return nil, errors.New("Unable to get a LinkRequest in Time")
}

// GetLinkRequest will check the dispatcher for a pending link request
func (c *Client) GetLinkRequest() (*PendingLinkRequest, error) {
	t := true
	msg := &RawMessage{
		Message: &wire.LinkTransfer{
			// I LOVE that ProtoBufs requires pointers
			Request: &t,
		},
		Code: wire.LinkTransferCode,
		Head: c.createHeader(c.Server),
	}

	resp, err := c.sendAndGetResponse(msg)
	if err != nil {
		return nil, err
	}

	enc, err := message.CreateEncryptedMessageFromBytes(resp.Data)
	if err != nil {
		return nil, err
	}

	signed, err := enc.Decrypt(c.Key)
	if err != nil {
		return nil, err
	}

	if !signed.Verify() {
		return nil, errors.New("Unable to verify signed message.")
	}

	data, typ, _, err := signed.ReconstructMessageWithTimestamp()
	if err != nil {
		return nil, err
	}

	if typ != wire.LinkRequestCode {
		return nil, errors.New("Wrong to type of code for message.")
	}

	unmarsh := &wire.LinkRequestPayload{}
	err = proto.Unmarshal(data, unmarsh)
	if err != nil {
		return nil, err
	}

	fingerprint := crypto.BytesToAddress(unmarsh.Signing)
	signing, err := crypto.BytesToKey(unmarsh.Signing)
	if err != nil {
		return nil, err
	}

	encrypting, err := crypto.BytesToRSA(unmarsh.Encrypting)
	if err != nil {
		return nil, err
	}

	return &PendingLinkRequest{
		From: &identity.Address{
			Fingerprint:   fingerprint,
			SigningKey:    signing,
			EncryptionKey: encrypting,
		},
		Verification: *unmarsh.Verification,
	}, nil
}

type VerificationCode string

func (c *LinkClient) GetVerificationCode() (VerificationCode, error) {
	max := big.NewInt(10)
	output := ""

	for i := 0; i < 6; i++ {
		number, err := rand.Int(rand.Reader, max)
		if err != nil {
			return VerificationCode(""), err
		}

		output += strconv.Itoa(int(number.Int64()))
	}

	return VerificationCode(output), nil
}

// LinkRequest sends a LinkRequest to the dispatcher
// for a specific account `to`
func (l *LinkClient) LinkRequest(verification VerificationCode) error {
	c := l.Client

	toAddr := l.User.String()
	verificationString := string(verification)

	payload := &RawMessage{
		Message: &wire.LinkRequestPayload{
			Encrypting:   crypto.RSAToBytes(l.Client.Key.Address.EncryptionKey),
			Signing:      crypto.KeyToBytes(l.Client.Key.Address.SigningKey),
			Verification: &verificationString,
		},
		Code: wire.LinkRequestCode,
		Head: c.createHeader(l.User),
	}

	signed, err := message.SignMessage(payload, c.Key)
	if err != nil {
		return err
	}

	enc, err := signed.EncryptWithKey(l.User)
	if err != nil {
		return err
	}

	bytes, err := enc.ToBytes()
	if err != nil {
		return err
	}

	msg := &RawMessage{
		Message: &wire.LinkData{
			Payload: bytes,
			For:     &toAddr,
		},
		Code: wire.LinkRequestCode,
		Head: c.createHeader(c.Server),
	}

	return c.sendAndCheck(msg)
}

// LinkAcceptRequest will upload the current identity
// to the dispatcher encrypted for `from`
func (c *Client) LinkAcceptRequest(request *PendingLinkRequest) error {
	toAddr := request.From.String()

	keyWriter := &bytes.Buffer{}
	_, err := c.Key.GobEncodeKey(keyWriter)
	if err != nil {
		return err
	}

	payload := &RawMessage{
		Message: &wire.LinkKeyPayload{
			Identity: keyWriter.Bytes(),
		},
		Code: wire.LinkKeyCode,
		Head: c.createHeader(request.From),
	}

	signed, err := message.SignMessage(payload, c.Key)
	if err != nil {
		return err
	}

	enc, err := signed.EncryptWithKey(request.From)
	if err != nil {
		return err
	}

	bytes, err := enc.ToBytes()
	if err != nil {
		return err
	}

	msg := &RawMessage{
		Message: &wire.LinkData{
			Payload: bytes,
			For:     &toAddr,
		},
		Code: wire.LinkKeyCode,
		Head: c.createHeader(c.Server),
	}

	return c.sendAndCheck(msg)
}

// LinkGetIdentity will download the identity from
// the dispatcher and decrypt it
func (l *LinkClient) LinkGetIdentity() (*identity.Identity, error) {
	c := l.Client

	t := true
	msg := &RawMessage{
		Message: &wire.LinkTransfer{
			// I LOVE that ProtoBufs requires pointers
			Approved: &t,
		},
		Code: wire.LinkTransferCode,
		Head: c.createHeader(c.Server),
	}

	resp, err := c.sendAndGetResponse(msg)
	if err != nil {
		return nil, err
	}

	if resp.Code == 12 {
		return nil, errors.New(resp.Message)
	}

	enc, err := message.CreateEncryptedMessageFromBytes(resp.Data)
	if err != nil {
		return nil, err
	}

	signed, err := enc.Decrypt(c.Key)
	if err != nil {
		return nil, err
	}

	if !signed.Verify() {
		return nil, errors.New("Unable to verify signed message.")
	}

	data, typ, header, err := signed.ReconstructMessageWithTimestamp()
	if err != nil {
		return nil, err
	}

	if header.From.String() != l.User.String() {
		return nil, errors.New("Link Transfer is from the wrong party.")
	}

	if typ != wire.LinkKeyCode {
		return nil, errors.New("Wrong to type of code for message.")
	}

	unmarsh := &wire.LinkKeyPayload{}
	err = proto.Unmarshal(data, unmarsh)
	if err != nil {
		return nil, err
	}

	return identity.GobDecodeKey(bytes.NewBuffer(unmarsh.Identity))
}
