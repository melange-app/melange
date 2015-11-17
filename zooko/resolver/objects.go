package resolver

import (
	"airdispat.ch/crypto"
	"airdispat.ch/identity"
)

// AppRegistration represents a registration for an application that
// has its manifest file stored at the AirDispatch URL in Manifest.
type AppRegistration struct {
	Name     string `json:"name"`
	Manifest string `json:"manifest"`
}

// ServerRegistration represents a registration for a server to be
// listed in the public directory with the given name and description.
type ServerRegistration struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	*Registration
}

// Registration reflects a JSON serializable registration to be
// included in the Namecoin blockchain.
type Registration struct {
	Address       string `json:"address"`
	EncryptionKey []byte `json:"encryption"`
	Location      string `json:"location"`
	Alias         string `json:"alias"`

	Redirects map[string]Redirect `json:"redirects,omitempty"`
}

func CreateRegistrationFromIdentity(id *identity.Identity, alias, location string) *Registration {
	serializedKey := crypto.RSAToBytes(id.Address.EncryptionKey)

	return &Registration{
		Address:       id.Address.String(),
		EncryptionKey: serializedKey,
		Location:      location,
		Alias:         alias,
	}
}

// Redirect allows lookups on registration types to be redirected to a
// different registration to find the details.
type Redirect struct {
	Alias   string `json:"alias"`
	Address string `json:"address"`
}
