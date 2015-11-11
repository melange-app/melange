package account

import (
	"encoding/gob"
	"io"
)

// Serialize will turn an account into a series of bytes ready to be
// written to disk.
func (a *Account) Serialize(w io.Writer) error {
	return gob.NewEncoder(w).Encode(a)
}

// CreateAccountFromBytes will deserialize an account from a series of
// bytes (possibly from the disk).
func CreateAccountFromReader(r io.Reader) (*Account, error) {
	account := new(Account)
	err := gob.NewDecoder(r).Decode(account)
	return account, err
}
