package router

import (
	"fmt"

	"airdispat.ch/crypto"
	"airdispat.ch/identity"
	"airdispat.ch/routing"
)

type Router struct {
	connectedToNetwork bool
	chain              *blockchainManager

	Key    *identity.Identity
	Server *identity.Address
}

func CreateRouter(connect bool) *Router {
	r := &Router{}

	if connect {
		r.ConnectToNetwork()
	}

	return r
}

func (r *Router) ConnectToNetwork() {
	r.chain = CreateBlockchainManager()

	for _, v := range BootstrapNodes {
		p, err := newPeer(v, r.chain)
		if err != nil {
			fmt.Println("Got error connecting to peer", err)
		}

		if err = p.loadHeaders(); err != nil {
			fmt.Println("Got error getting headers", err)
		}
	}

	r.connectedToNetwork = true
}

func (r *Router) NamecoinLookup(addr string) ([]byte, error) {
	return r.lookupZookoServer(addr, name, 0)
}

func (r *Router) Lookup(addr string, name routing.LookupType) (*identity.Address, error) {
	panic("Cannot call lookup on Namecoin Router.")
}

func (r *Router) LookupAlias(addr string, name routing.LookupType) (*identity.Address, error) {
	// return r.lookupDNSChain(addr, name, 0)
	return r.lookupAddressZookoServer(addr, name, 0)
}

type dnsChainResult struct {
	Version string            `json:"version"`
	Header  map[string]string `json:"header"`
	Data    struct {
		Name      string                `json:"name"`
		Value     *namecoinRegistration `json:"value"`
		TXID      string                `json:"txid"`
		Address   string                `json:"address"`
		ExpiresIn int                   `json:"expires_in"`
		Expired   int                   `json:"expired"`
	} `json:"data"`
}

type namecoinRegistration struct {
	Name          string                      `json:"name"`
	SigningKey    []byte                      `json:"signing_key"`
	EncryptionKey []byte                      `json:"encryption_key"`
	Location      string                      `json:"location"`
	Redirects     map[string]routing.Redirect `json:"redirects"`
}

func (n *namecoinRegistration) BuildAddress() (*identity.Address, error) {
	signing, err := crypto.BytesToKey(n.SigningKey)
	if err != nil {
		return nil, err
	}

	encryption, err := crypto.BytesToRSA(n.EncryptionKey)
	if err != nil {
		return nil, err
	}

	return &identity.Address{
		Alias:         n.Name,
		EncryptionKey: encryption,
		SigningKey:    signing,
		Location:      n.Location,
		Fingerprint:   crypto.BytesToAddress(n.SigningKey),
	}, nil
}

/*
{
  "name": "hleath",
  "signing_key": [],
  "encryption_key": [],
  "redirects": {
     "TXMail": {
        "fingerprint": "5523sdf...",
        "alias": "hleath_server",
     }
  }
}
*/

func (r *Router) Register(id *identity.Identity, alias string, redirects map[string]routing.Redirect) error {
	// TODO: hleath
	return nil
}
