package models

import (
	"airdispat.ch/identity"
	"github.com/huntaub/go-db"
)

// Identity represents a Keypair
type Identity struct {
	Id          db.PrimaryKey
	Nickname    string
	Fingerprint string
	Server      string
	// Actual Data
	Data []byte `json:"-"`
	// Password Protection
	Protected bool
	Aliases   *db.HasMany `table:"alias" on:"identity" json:"-"`
}

func CreateIdentity(nick string, id *identity.Identity, password string) *Identity {
	return nil
}

// Alias represent a registered Identity
type Alias struct {
	Id       db.PrimaryKey
	Identity *db.HasOne `table:"identity"`
	Location string
	Username string
}
