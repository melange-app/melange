package account

import (
	"github.com/melange-app/nmcd/btcutil"
	"github.com/melange-app/nmcd/chaincfg"
)

// getPubKeyHash will return the public key hash of the current
// account.
func (a *Account) getPubKeyHash() (*btcutil.AddressPubKeyHash, error) {
	compressed := a.Keys.PubKey().SerializeCompressed()

	return btcutil.NewAddressPubKeyHash(btcutil.Hash160(compressed), &chaincfg.MainNetParams)
}
