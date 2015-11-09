package account

import (
	"github.com/melange-app/nmcd/btcec"
	"github.com/melange-app/nmcd/btcutil"
)

// Account represents a Namecoin account that does not see the
// blockchain. Therefore, it must store its list of UTXO internally.
type Account struct {
	Keys    *btcec.PrivateKey
	Unspent UTXOList
}

// CreateAccount will generate a new Namecoin address and associated
// account.
func CreateAccount() (*Account, error) {
	key, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}

	return &Account{
		Keys: key,
	}, nil
}

// RequestFunds will ask the local Zooko server to send money to
// register a new name.
func (a *Account) RequestFunds() (*UTXO, error) {
	return nil, nil
}

// Balance will return the balance of the wallet.
func (a *Account) Balance() int64 {
	return a.Unspent.Balance()
}

// We should only ever need the key associated with the account. If we
// need to sign for something else, this will not work.
func (a *Account) GetKey(btcutil.Address) (*btcec.PrivateKey, bool, error) {
	return a.Keys, true, nil
}

// UTXOList represents the list of transactions that are unspent. We
// create a separate type so that we can satisfy the sort interface.
type UTXOList []*UTXO

func (a UTXOList) Less(i, j int) bool { return a[i].Amount > a[j].Amount }
func (a UTXOList) Len() int           { return len(a) }
func (a UTXOList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// Balance will tell you how much Namecoin is in a specific set of
// unspent transactions.
func (a UTXOList) Balance() int64 {
	var balance int64
	for _, v := range a {
		balance += v.Amount
	}

	return balance
}

// UTXO represents an unspent transaction output.
type UTXO struct {
	TxID     string
	Output   uint32
	Amount   int64
	PkScript []byte
}
