package packaging

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"airdispat.ch/crypto"
	"airdispat.ch/identity"
)

// Provider is a JSON Object that represents either a Server or a Tracker.
// It is used to pass information from getmelange.com to the Melange client.
type Provider struct {
	Id            string            `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Image         string            `json:"image"`
	URL           string            `json:"url"`
	Fingerprint   string            `json:"fingerprint"`
	EncryptionKey string            `json:"encryption_key"`
	Proof         string            `json:"proof"`
	Users         int               `json:"users"`
	Key           *identity.Address `json:"-"`
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

// http://www.getmelange.com/api
// GET /servers
// GET /trackers
// GET /applications

// Packager is an object that returns providers from getmelange.com
type Packager struct {
	API   string
	cache map[string]map[string]*Provider
}

func (p *Packager) DecodeProviders(url string) ([]*Provider, error) {
	if p.cache == nil {
		p.cache = make(map[string]map[string]*Provider)
	}

	if p.cache[url] == nil {
		p.cache[url] = make(map[string]*Provider)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		dec := json.NewDecoder(resp.Body)

		var obj []*Provider
		err = dec.Decode(&obj)
		if err != nil {
			return nil, err
		}

		var out = make([]*Provider, 0)
		for _, v := range obj {
			err := v.LoadDefaults()
			if err == nil {
				out = append(out, v)
				p.cache[url][v.Id] = v
			}
		}
		return out, nil
	}

	var out = make([]*Provider, len(p.cache[url]))
	i := 0
	for _, v := range p.cache[url] {
		out[i] = v
		i++
	}
	return out, nil
}

// GetServers will download Servers from getmelange.com.
func (p *Packager) GetServers() ([]*Provider, error) {
	return p.DecodeProviders(p.API + "/servers")
}

// GetTrackers will download Trackers from getmelange.com.
func (p *Packager) GetTrackers() ([]*Provider, error) {
	return p.DecodeProviders(p.API + "/trackers")
}

func (p *Packager) getFromId(id string, url string) (*Provider, error) {
	if p.cache == nil || p.cache[url] == nil {
		_, err := p.DecodeProviders(url)
		if err != nil {
			return nil, err
		}
	}
	provider, ok := p.cache[url][id]
	if !ok {
		return nil, errors.New("That id doesn't exist.")
	}
	return provider, nil
}

// TrackerFromId will return the Provider instance associated with a particular
// Id.
func (p *Packager) TrackerFromId(id string) (*Provider, error) {
	return p.getFromId(id, p.API+"/trackers")
}

// ServerFromId will return the Provider instance associated with a particular
// Id.
func (p *Packager) ServerFromId(id string) (*Provider, error) {
	return p.getFromId(id, p.API+"/servers")
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
