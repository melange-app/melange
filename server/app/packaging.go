package app

import (
	"airdispat.ch/crypto"
	"airdispat.ch/identity"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

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

type Packager struct {
	API   string
	cache map[string]map[string]*Provider
}

func (p *Packager) decodeProviders(url string) ([]*Provider, error) {
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
	} else {
		var out = make([]*Provider, len(p.cache[url]))
		i := 0
		for _, v := range p.cache[url] {
			out[i] = v
			i++
		}
		return out, nil
	}
}

func (p *Packager) GetServers() ([]*Provider, error) {
	return p.decodeProviders(p.API + "/servers")
}

func (p *Packager) GetTrackers() ([]*Provider, error) {
	return p.decodeProviders(p.API + "/trackers")
}

func (p *Packager) getFromId(id string, url string) (*Provider, error) {
	if p.cache == nil || p.cache[url] == nil {
		_, err := p.decodeProviders(url)
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

func (p *Packager) TrackerFromId(id string) (*Provider, error) {
	return p.getFromId(id, p.API + "/trackers")
}

func (p *Packager) ServerFromId(id string) (*Provider, error) {
	return p.getFromId(id, p.API + "/servers")
}

func (p *Packager) GetApps() ([]*struct{}, error) {
	return nil, nil
}

func (p *Packager) AppFromId(id string) (*struct{}, error) {
	return nil, nil
}
