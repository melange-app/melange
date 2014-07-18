package models

import ()

// Identity represents a Keypair
type Identity struct {
	Nickname    string
	Fingerprint string
	// Actual Data
	Encryption []byte
	Signing    []byte
	// Password Protection
	Protected bool
}

// Alias represent a registered Identity
type Alias struct {
	IdentityId int
	Location   string
	Username   string
}
