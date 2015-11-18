package packaging

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"getmelange.com/zooko/resolver"

	"airdispat.ch/crypto"
	"airdispat.ch/identity"
)

const (
	melangeServerPrefix = "mlg/server/"
)

// Provider is a JSON Object that represents either a Server or a Tracker.
// It is used to pass information from getmelange.com to the Melange client.
type Provider struct {
	Id string `json:"id"`
	// Metadata
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	URL         string `json:"url"`
	// Addressing Properties
	Alias         string `json:"alias"`
	Fingerprint   string `json:"fingerprint"`
	EncryptionKey string `json:"encryption_key"`
	// Random
	Proof string            `json:"proof"`
	Users int               `json:"users"`
	Key   *identity.Address `json:"-"`
}

func CreateProviderFromRegistration(r *resolver.Registration) (*Provider, error) {
	address := identity.CreateAddressFromString(r.Address)
	key, err := crypto.BytesToRSA(r.EncryptionKey)
	if err != nil {
		return nil, err
	}

	address.EncryptionKey = key
	address.Location = r.Location

	return &Provider{
		Id: melangeServerPrefix + r.Alias,

		// Filler for when we actually have market data.
		Name:        r.Alias,
		Description: r.Alias,

		// Information needed to send to the server
		Fingerprint:   r.Address,
		EncryptionKey: hex.EncodeToString(r.EncryptionKey),
		Alias:         r.Alias,
		URL:           r.Location,

		// The AirDispatch address that we really care about.
		Key: address,
	}, nil
}

// LoadDefaults will do several things on a provider.
// 1. Create an ID Based on its Name, URL, and Fingerprint
// 2. Set a default image for the Melange Client.
// 3. Create the (*identity).Address
func (p *Provider) LoadDefaults() error {
	p.Id = fmt.Sprintf("%s-%s-%s", strings.ToLower(p.Name), strings.ToLower(p.URL), strings.ToLower(p.Fingerprint))
	p.Image = "http://placehold.it/400"

	p.Key = identity.CreateAddressFromString(p.Fingerprint)
	data, err := hex.DecodeString(p.EncryptionKey)
	if err != nil {
		return err
	}
	key, err := crypto.BytesToRSA(data)
	if err != nil {
		return err
	}
	p.Key.EncryptionKey = key
	p.Key.Location = p.URL
	return nil
}

// Packager will return the Melange resources from the blockchain.
type Packager struct {
	API    string
	Plugin string
	Debug  bool
	cache  map[string]map[string]*Provider
}

// GetServers will download Servers from getmelange.com.
func (p *Packager) GetServers() ([]*Provider, error) {
	id, err := identity.CreateIdentity()
	if err != nil {
		return nil, err
	}

	client := resolver.CreateClient(id)
	names, err := client.LookupPrefix(melangeServerPrefix)
	if err != nil {
		return nil, err
	}

	var providers []*Provider
	for _, name := range names {
		registration, found, err := client.Lookup(name)
		if err != nil || !found {
			// Ignore errors from objects that cannot
			// return their registrations.
			continue
		}

		provider, err := CreateProviderFromRegistration(registration)
		if err != nil {
			// Ignore errors about incorrect
			// serialization. They will not be included in
			// the marketplace.
			continue
		}

		providers = append(providers, provider)
	}

	return providers, nil
}

// ServerFromId will return the Provider instance associated with a particular
// Id.
func (p *Packager) ServerFromId(id string) (*Provider, error) {
	// Ensure that we are looking up a server.
	if !strings.HasPrefix(id, melangeServerPrefix) {
		id = melangeServerPrefix + id
	}

	key, err := identity.CreateIdentity()
	if err != nil {
		return nil, err
	}

	registration, found, err := resolver.CreateClient(key).Lookup(id)
	if err != nil {
		return nil, err
	} else if !found {
		return nil, errors.New("packaging: cannot get a server that doesn't exist")
	}

	return CreateProviderFromRegistration(registration)
}

// GetApps will (in the future) return the Applications from getmelange.com
func (p *Packager) GetApps() ([]*struct{}, error) {
	return nil, nil
}

// AppFromId will (in the future) return the Application associated with a
// particular id.
func (p *Packager) AppFromId(id string) (*struct{}, error) {
	return nil, nil
}
