package account

import (
	"errors"

	"github.com/melange-app/nmcd/btcec"
	"github.com/melange-app/nmcd/btcutil"
	"github.com/melange-app/nmcd/chaincfg"
)

// Key is a type alias here so that we can force gob to utilize the
// Serialization methods that btcec provides rather than writing our
// own.
type Key btcec.PrivateKey

// Key will return the btcec representation of the key.
func (k *Key) Key() *btcec.PrivateKey {
	return (*btcec.PrivateKey)(k)
}

// GobEncode will encode the key using the Serialize function.
func (k *Key) GobEncode() ([]byte, error) {
	return (*btcec.PrivateKey)(k).Serialize(), nil
}

// GobDecode will decode the key using the PrivKeyFromBytes function.
func (k *Key) GobDecode(b []byte) error {
	key, _ := btcec.PrivKeyFromBytes(btcec.S256(), b)
	if key == nil {
		return errors.New("zooko: couldn't unmarshal key")
	}

	*k = Key(*key)
	return nil
}

// getPubKeyHash will return the public key hash of the current
// account.
func (a *Account) PublicKeyHash() (*btcutil.AddressPubKeyHash, error) {
	compressed := a.Keys.Key().PubKey().SerializeCompressed()

	return btcutil.NewAddressPubKeyHash(btcutil.Hash160(compressed), &chaincfg.MainNetParams)
}
