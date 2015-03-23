package models

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"airdispat.ch/crypto"
	"airdispat.ch/identity"
	gdb "github.com/huntaub/go-db"
)

// Identity represents a Keypair
type Identity struct {
	Id          gdb.PrimaryKey `json:"-"`
	Nickname    string
	Fingerprint string

	// Server Information and Tacking
	Server            string
	ServerKey         string
	ServerFingerprint string
	ServerAlias       string

	// Actual Data
	Data []byte `json:"-"`

	// Password Protection
	Protected bool

	// Alias Support
	Aliases       *gdb.HasMany `table:"alias" on:"identity" json:"-"`
	LoadedAliases []*Alias     `db:"-"`

	// Profile
	Profile *gdb.HasOne `table:"profile"`

	// Transcendence
	Current bool `db:"-"`
}

// CreateIdentityFromDispatch will take an (*identity).Identity from the AirDispatch
// package and a key (to encrypt the identity with) and return an *Identity
// suitable for inserting into the database.
func CreateIdentityFromDispatch(id *identity.Identity, key string) (*Identity, error) {
	buffer := &bytes.Buffer{}
	_, err := id.GobEncodeKey(buffer)
	if err != nil {
		return nil, err
	}

	return &Identity{
		Fingerprint: id.Address.String(),
		Data:        buffer.Bytes(),
		Protected:   key != "",
	}, nil
}

// ToDispatch will take a key (to decode the identity) and return an (*identity).Identity
// suitable for using in AirDispatch-related objects.
func (i *Identity) ToDispatch(s gdb.Executor, key string) (*identity.Identity, error) {
	data := i.Data
	if key != "" {
		return nil, errors.New("We don't support encryption yet.")
	}

	buf := bytes.NewBuffer(data)

	id, err := identity.GobDecodeKey(buf)
	if err != nil {
		return nil, err
	}

	id.SetLocation(i.Server)

	// We sort of just pick one randomly here.

	alias := &Alias{}
	err = i.Aliases.One(s, alias)
	if err != nil {
		return nil, err
	}

	id.Address.Alias = alias.String()

	return id, nil
}

func (id *Identity) CreateServerFromIdentity() (*identity.Address, error) {
	data, err := hex.DecodeString(id.ServerKey)
	if err != nil {
		return nil, err
	}

	key, err := crypto.BytesToRSA(data)
	if err != nil {
		return nil, err
	}

	fingerprint, err := hex.DecodeString(id.ServerFingerprint)
	if err != nil {
		return nil, err
	}

	return &identity.Address{
		Fingerprint:   fingerprint,
		Location:      id.Server,
		EncryptionKey: key,
		Alias:         id.ServerAlias,
	}, nil
}

// Alias represent a registered Identity
type Alias struct {
	Id       gdb.PrimaryKey `json:"-"`
	Identity *gdb.HasOne    `json:"-"`
	Location string
	Username string
}

func (a *Alias) String() string {
	if a.Location != "_namecoin" {
		return fmt.Sprintf("%s@%s", a.Username, a.Location)
	}
	return fmt.Sprintf("/%s", a.Username)
}

type Profile struct {
	Id          gdb.PrimaryKey `json:"-"`
	Name        string         `json:"name"`
	Image       string         `json:"image"`
	Description string         `json:"description"`
}
