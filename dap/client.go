package dap

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"getmelange.com/dap/wire"

	"crypto/rand"

	"airdispat.ch/crypto"
	adErrors "airdispat.ch/errors"
	"airdispat.ch/identity"
	"airdispat.ch/message"
	adWire "airdispat.ch/wire"
)

func (c *Client) checkForError(d []byte, typ string, h message.Header) *adErrors.Error {
	if typ != adWire.ErrorCode {
		return nil
	}

	return adErrors.CreateErrorFromBytes(d, h)
}

type Client struct {
	Key    *identity.Identity
	Server *identity.Address
}

func CreateClient(key *identity.Identity, server *identity.Address) *Client {
	if server.Location == "" {
		panic("Cannot use Server without an address component.")
	}

	return &Client{key, server}
}

func (c *Client) createHeader(to ...*identity.Address) message.Header {
	header := message.CreateHeader(c.Key.Address, to...)
	header.EncryptionKey = crypto.RSAToBytes(c.Key.Address.EncryptionKey)
	return header
}

func (c *Client) sendAndGetResponse(msg message.Message) (*Response, error) {
	data, typ, head, err := message.SendMessageAndReceiveWithTimestamp(msg, c.Key, c.Server)

	if err != nil {
		return nil, err
	} else if adErr := c.checkForError(data, typ, head); adErr != nil {
		return nil, adErr
	} else if typ != wire.ResponseCode {
		return nil, errors.New("Unexpected response type:" + typ)
	}

	return CreateResponseFromBytes(data, head)
}

func (c *Client) sendAndCheck(msg message.Message) error {
	response, err := c.sendAndGetResponse(msg)
	if err != nil {
		return err
	}
	if response.Code != 0 {
		return errors.New("Unexpected Return Type")
	}
	return nil
}

// -------
// Account Management
// -------

// Register with the Server
func (c *Client) Register(keys map[string][]byte) error {
	outKeys := make([]*wire.Data, 0)
	for key, value := range keys {
		newKey := key
		outKeys = append(outKeys, &wire.Data{
			Key:  &newKey,
			Data: value,
		})
	}

	msg := &RawMessage{
		Message: &wire.Register{
			Keys: outKeys,
		},
		Code: wire.RegisterCode,
		Head: c.createHeader(c.Server),
	}

	return c.sendAndCheck(msg)
}

// Unregister with the Server
func (c *Client) Unregister(keys map[string][]byte) error {
	outKeys := make([]*wire.Data, 0)
	for key, value := range keys {
		newKey := key
		outKeys = append(outKeys, &wire.Data{
			Key:  &newKey,
			Data: value,
		})
	}

	msg := &RawMessage{
		Message: &wire.Unregister{
			Keys: outKeys,
		},
		Code: wire.UnregisterCode,
		Head: c.createHeader(c.Server),
	}

	return c.sendAndCheck(msg)
}

// -------
// Message Management
// -------

func (c *Client) decryptAndVerify(msg *message.EncryptedMessage, ts bool) ([]byte, string, message.Header, error) {
	receivedSign, err := msg.Decrypt(c.Key)
	if err != nil {
		return nil, "", message.Header{}, err
	}

	if !receivedSign.Verify() {
		return nil, "", message.Header{}, errors.New("Unable to Verify Message")
	}

	if ts {
		return receivedSign.ReconstructMessageWithTimestamp()
	} else {
		return receivedSign.ReconstructMessage()
	}
}

func (c *Client) signAndEncrypt(msg message.Message, to ...*identity.Address) ([]byte, error) {
	enc, err := c.signAndEncryptMessage(msg, to...)
	if err != nil {
		return nil, err
	}

	return enc.ToBytes()
}

func (c *Client) signAndEncryptMessage(msg message.Message, to ...*identity.Address) (*message.EncryptedMessage, error) {
	signed, err := message.SignMessage(msg, c.Key)
	if err != nil {
		return nil, err
	}

	if len(to) == 0 {
		return signed.UnencryptedMessage(nil)
	}

	enc, err := signed.EncryptWithKey(to[0])
	if err != nil {
		return nil, err
	}

	for i, v := range to {
		if i != 0 {
			err = enc.AddRecipient(v)
			if err != nil {
				return nil, err
			}
		}
	}

	return enc, nil
}

// Download Messages from Server
func (c *Client) DownloadSentMessages(since uint64, context bool) ([]*ResponseMessage, error) {
	return c.downloadMessages(since, context, true)
}

func (c *Client) DownloadMessages(since uint64, context bool) ([]*ResponseMessage, error) {
	return c.downloadMessages(since, context, false)
}

func (c *Client) downloadMessages(since uint64, context bool, sent bool) ([]*ResponseMessage, error) {
	msg := &RawMessage{
		Message: &wire.DownloadMessages{
			Since:   &since,
			Context: &context,
			Sent:    &sent,
		},
		Code: wire.DownloadMessagesCode,
		Head: c.createHeader(c.Server),
	}

	conn, err := message.ConnectToServer(c.Server.Location)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = message.SignAndSendToConnection(msg, c.Key, c.Server, conn)
	if err != nil {
		return nil, err
	}

	responseContainer, err := message.ReadMessageFromConnection(conn)
	if err != nil {
		return nil, err
	}

	data, typ, head, err := c.decryptAndVerify(responseContainer, true)
	if err != nil {
		return nil, err
	} else if adErr := c.checkForError(data, typ, head); adErr != nil {
		return nil, adErr
	} else if typ != wire.ResponseCode {
		return nil, errors.New("Unexpected response type:" + typ + " expected " + wire.ResponseCode)
	}

	response, err := CreateResponseFromBytes(data, head)
	if err != nil {
		return nil, err
	}

	if response.Length == 0 {
		return nil, nil
	}

	messages := make([]*ResponseMessage, response.Length)
	// Read the returned messages into an array.
	for i := uint64(0); i < response.Length; i++ {
		responseContainer, err := message.ReadMessageFromConnection(conn)
		if err != nil {
			return nil, err
		}

		data, typ, head, err := c.decryptAndVerify(responseContainer, true)
		if err != nil {
			if sent {
				fmt.Println("Downloaded a message that does not container sender decryption.")
				continue
			} else {
				return nil, err
			}
		} else if adErr := c.checkForError(data, typ, head); adErr != nil {
			return nil, adErr
		} else if typ != wire.ResponseMessageCode {
			return nil, errors.New("Unexpected response type:" + typ + " expected " + wire.ResponseMessageCode)
		}

		rspMsg, err := CreateResponseMessageFromBytes(data, head)
		if err != nil {
			return nil, err
		}

		messages[i] = rspMsg
	}
	return messages, nil
}

// Publish Message on Server
func (c *Client) PublishMessage(enc *message.Mail, to []*identity.Address, name string, alert bool) (string, error) {
	// Add the sender to the recipient list so that they can decrypt it
	// later during downloading of messages.
	bytes, err := c.signAndEncrypt(enc, append(to, c.Key.Address)...)
	if err != nil {
		return "", err
	}

	// If name is blank, use the hash value of the message as the name.
	if name == "" {
		hash := crypto.HashSHA(bytes)
		name = hex.EncodeToString(hash)
	}

	toAddrs := make([]string, len(to))
	for i, v := range to {
		toAddrs[i] = v.String()
	}

	msg := &RawMessage{
		Message: &wire.PublishMessage{
			To:    toAddrs,
			Name:  &name,
			Alert: &alert,
			Data:  bytes,
		},
		Code: wire.PublishMessageCode,
		Head: c.createHeader(c.Server),
	}

	return name, c.sendAndCheck(msg)
}

// Update Message on Server
func (c *Client) UpdateMessage(enc *message.Mail, to []*identity.Address, name string) error {
	bytes, err := c.signAndEncrypt(enc, to...)
	if err != nil {
		return err
	}

	if name == "" {
		return errors.New("A named message to update is required.")
	}

	msg := &RawMessage{
		Message: &wire.UpdateMessage{
			Name: &name,
			Data: bytes,
		},
		Code: wire.UpdateMessageCode,
		Head: c.createHeader(c.Server),
	}

	return c.sendAndCheck(msg)
}

// ----
// AD Data
// ----

func (c *Client) PublishDataMessage(
	r io.ReadSeeker,
	to []*identity.Address,
	typ, prefix, file string) (string, error) {
	// Hash the Plaintext
	hasher := sha256.New()

	// Read Once
	n, err := io.Copy(hasher, r)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Return to Beginning
	_, err = r.Seek(0, 0)
	if err != nil {
		return "", err
	}

	// Draw a Random Key
	key := make([]byte, 32)
	rand.Read(key)

	length := uint64(n) + uint64(aes.BlockSize)

	name := fmt.Sprintf("%s:%x", prefix, hasher.Sum(nil))

	// Sign and Encrypt the Data Header
	encData, err := c.signAndEncrypt(
		&RawMessage{
			Message: &adWire.Data{
				Hash:   hasher.Sum(nil),
				Length: &length,
				Key:    key,
				Type:   &typ,
				Name:   &name,
				File:   &file,
			},
			Code: adWire.DataCode,
			Head: c.createHeader(to...),
		}, to...)
	if err != nil {
		return "", err
	}

	// Get the Encrypted Hash
	encHasher := sha256.New()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create the CFB IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	encHasher.Write(iv)

	stream := cipher.NewCFBEncrypter(block, iv)

	// Create the StreamWriter to Write Encrypted to the Hash Function
	encWriter := cipher.StreamWriter{
		S: stream,
		W: encHasher,
	}

	// Read Twice
	_, err = io.Copy(encWriter, r)
	if err != nil {
		return "", err
	}

	// Return to Beginning
	_, err = r.Seek(0, 0)
	if err != nil {
		return "", err
	}

	// Write PublishDataMessage

	// Connect to Server
	conn, err := message.ConnectToServer(c.Server.Location)
	if err != nil {
		return "", err
	}

	var toAddrs []string
	for _, v := range to {
		toAddrs = append(toAddrs, v.String())
	}

	// Create the Encrypted Message
	enc, err := c.signAndEncryptMessage(&RawMessage{
		Message: &wire.PublishDataMessage{
			Header: encData,
			Hash:   encHasher.Sum(nil),
			Name:   &name,
			Length: &length,
			To:     toAddrs,
		},
		Code: wire.PublishDataMessageCode,
		Head: c.createHeader(c.Server),
	}, c.Server)
	if err != nil {
		return "", err
	}

	err = enc.SendMessageToConnection(conn)
	if err != nil {
		return "", err
	}

	_, err = conn.Write(iv)
	if err != nil {
		return "", err
	}

	// Write actual Data
	actualStream := cipher.StreamWriter{
		S: cipher.NewCFBEncrypter(block, iv),
		W: conn,
	}

	// Three Times, You're Out
	_, err = io.Copy(actualStream, r)
	if err != nil {
		return "", err
	}

	return name, adErrors.CheckConnectionForError(conn)
}

// ----
// Data Management
// ----

// Set Arbitrary Data on Server
func (c *Client) SetData(key string, value []byte) error {
	msg := &RawMessage{
		Message: &wire.Data{
			Key:  &key,
			Data: value,
		},
		Code: wire.DataCode,
		Head: c.createHeader(c.Server),
	}

	return c.sendAndCheck(msg)
}

func (c *Client) GetData(key string) ([]byte, error) {
	msg := &RawMessage{
		Message: &wire.GetData{
			Key: &key,
		},
		Code: wire.GetDataCode,
		Head: c.createHeader(c.Server),
	}

	resp, err := c.sendAndGetResponse(msg)
	if err != nil {
		return nil, err
	}

	return resp.Data, err
}
