package account

import (
	"github.com/melange-app/nmcd/btcec"
	"github.com/melange-app/nmcd/btcutil"
)

const (
	// NameMinimumBalance is the minimum balance that an account
	// can have in order to perform a name operation.
	NameMinimumBalance = txFee + nameNetworkFee
)

// Account represents a Namecoin account that does not see the
// blockchain. Therefore, it must store its list of UTXO internally.
type Account struct {
	Keys    *Key
	Unspent UTXOList
	Pending []*Transaction
}

// CreateAccount will generate a new Namecoin address and associated
// account.
func CreateAccount() (*Account, error) {
	key, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	accountKey := Key(*key)

	return &Account{
		Keys: &accountKey,
	}, nil
}

// Commit will update the UTXO graph for this account to reflect the
// changes in Transaction.
func (a *Account) Commit(t *Transaction) {
	var utxo UTXOList

	// Add all the new transactions to the list.
	for _, v := range t.New {
		utxo = append(utxo, v)
	}

	// Add all the transactions that aren't in the spent list to
	// the array.
	for _, v := range a.Unspent {
		for _, x := range t.Spent {
			if !v.Equals(x) {
				utxo = append(utxo, v)
			}
		}
	}

	a.Unspent = utxo
}

// Balance will return the balance of the wallet.
func (a *Account) Balance() int64 {
	return a.Unspent.Balance()
}

// We should only ever need the key associated with the account. If we
// need to sign for something else, this will not work.
func (a *Account) GetKey(btcutil.Address) (*btcec.PrivateKey, bool, error) {
	return a.Keys.Key(), true, nil
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

func (u *UTXO) Equals(y *UTXO) bool {
	return (u.TxID == y.TxID) && (u.Output == y.Output)
}
