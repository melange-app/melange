package message

import (
	"airdispat.ch/crypto"
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"code.google.com/p/goprotobuf/proto"
)

// namecoinTransaction is used when the message must carry proof code
// that the transactions are in the Namecoin blockchain.
type namecoinTransaction struct {
	TxId               string
	Branch             int32
	VerificationHashes [][]byte
	Raw                []byte
	BlockId            string
	Value              int32
}

type rawMessage struct {
	header      message.Header
	messageType string
	data        []byte
}

// CreateRawMessage will build a message that can be used with
// AirDispatch out of the protocol buffers that Zooko is using.
func CreateMessage(
	m proto.Message,
	from *identity.Identity,
	to ...*identity.Address) (message.Message, error) {
	data, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}

	typ := getMessageType(m)

	hdr := message.CreateHeader(from.Address, to...)
	hdr.EncryptionKey = crypto.RSAToBytes(from.Address.EncryptionKey)

	return &rawMessage{
		data:        data,
		messageType: typ,
		header:      hdr,
	}, nil
}

// These methods satisfy the message.Message interface.
func (r *rawMessage) Type() string           { return r.messageType }
func (r *rawMessage) Header() message.Header { return r.header }
func (r *rawMessage) ToBytes() []byte        { return r.data }

// Compile-time check that our messae satisfies the message.Message
// interface.
var _ message.Message = &rawMessage{}
