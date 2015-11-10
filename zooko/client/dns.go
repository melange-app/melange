package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"airdispat.ch/identity"
	"airdispat.ch/routing"
)

func (r *Router) lookupDNSChain(addr string, name routing.LookupType, redirects int) (*identity.Address, error) {
	if redirects > 10 {
		return nil, errors.New("zooko: too many redirects")
	}

	url := fmt.Sprintf("http://api.dnschain.net/ad/%s", addr)
	fmt.Println("Using URL:", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	} else if resp.StatusCode == 404 || resp.StatusCode == 400 {
		return nil, errors.New("zooko: name not found")
	}

	var d dnsChainResult
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, err
	}
	resp.Body.Close()

	if d.Data.Expired == 1 {
		return nil, errors.New("zooko: that name has expired")
	}

	if redirect, ok := d.Data.Value.Redirects[string(name)]; ok {
		addr, err := r.lookupDNSChain(redirect.Alias, name, redirects+1)
		if err != nil {
			return nil, err
		}

		if addr.String() != redirect.Fingerprint {
			return nil, errors.New("zooko: redirect fingerprints do not match")
		}

		return addr, nil
	}

	return d.Data.Value.BuildAddress()
}
