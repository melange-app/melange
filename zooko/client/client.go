package router

import (
	"errors"
	"fmt"

	"airdispat.ch/crypto"
	adErrors "airdispat.ch/errors"
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"
	"airdispat.ch/wire"
	zmessage "getmelange.com/zooko/message"
)

func (r *Router) lookupAddressZookoServer(addr string, name routing.LookupType, redirects int) (*identity.Address, error) {
	data, err := r.lookupZookoServer(addr)
	if err != nil {
		return nil, err
	}
}

func (r *Router) lookupZookoServer(addr string) ([]byte, error) {
	lookup := &zmessage.LookupNameMessage{
		Name: addr,
		H:    message.CreateHeader(r.Key.Address, r.Server),
	}

	// This needs to be handled by the standard library.
	lookup.H.EncryptionKey = crypto.RSAToBytes(r.Key.Address.EncryptionKey)

	data, typ, h, err := message.SendMessageAndReceiveWithTimestamp(lookup, r.Key, r.Server)
	if err != nil {
		return nil, err
	} else if typ == wire.ErrorCode {
		return nil, adErrors.CreateErrorFromBytes(data, h)
	} else if typ != zmessage.ResolvedNameCode {
		return nil, errors.New("zooko: message received from server is of incorrect type")
	}

	rn, err := zmessage.CreateResolvedNameMessageFromBytes(data, h)
	if err != nil {
		return nil, err
	}

	fmt.Println("Received", len(rn.Transactions), "transactions worth of information.")

	if verify := r.chain.VerifyTransactions(rn.Transactions...); !verify {
		return nil, errors.New("zooko: returned transactions are not in blokcchain")
	}

	return nil, nil
}
