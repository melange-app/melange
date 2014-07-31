package router

// This file describes two routes that are used throughout the Melange System.

import (
	"errors"
	"fmt"
	"strings"

	"airdispat.ch/identity"
	"airdispat.ch/routing"
	"airdispat.ch/tracker"
	cache "github.com/huntaub/go-cache"
)

var routeCache *cache.Cache

// var cLock sync.RWMutex

var ServerKey *identity.Identity

type Router struct {
	Origin      *identity.Identity
	TrackerList []string
	Redirects   int
}

func (r *Router) HandleRedirect(name routing.LookupType, redirect routing.Redirect) (*identity.Address, error) {
	r.Redirects++
	if r.Redirects > 10 {
		return nil, errors.New("Too many redirects.")
	}
	data, err := r.LookupAlias(redirect.Alias, name)
	if err != nil {
		return nil, err
	}
	if data.String() != redirect.Fingerprint {
		return nil, errors.New("Redirected data doesn't have the same fingerprint.")
	}
	return data, nil
}

func (a *Router) LookupAlias(from string, name routing.LookupType) (*identity.Address, error) {
	test, stale := routeCache.Get(from)
	if !stale {
		return test.(*identity.Address), nil
	}

	if from[0] == '/' {
		return a.Lookup(from[1:], name)
	}
	comp := strings.Split(from, "@")
	if len(comp) != 2 {
		return nil, errors.New("Can't use lookup router without tracker address.")
	}

	url := tracker.GetTrackingServerLocationFromURL(comp[1])
	t := &tracker.Router{
		URL:        url,
		Origin:     a.Origin,
		Redirector: a,
	}

	addr, err := t.LookupAlias(comp[0], name)
	if err == nil {
		routeCache.Store(from, addr)
		routeCache.Store(addr.String(), addr)
		return addr, nil
	}
	return nil, err
}

func (a *Router) Lookup(from string, name routing.LookupType) (*identity.Address, error) {
	test, stale := routeCache.Get(from)
	if !stale {
		return test.(*identity.Address), nil
	}

	for _, v := range a.TrackerList {
		fmt.Println(v, ServerKey)
		a, err := (&tracker.Router{
			URL:        v,
			Origin:     ServerKey,
			Redirector: a,
		}).Lookup(from, name)
		if err == nil {
			routeCache.Store(from, a)
			routeCache.Store(a.String(), a)
			return a, nil
		}
	}
	return nil, errors.New("Couldn't find address in Trackers.")
}

func (a *Router) Register(key *identity.Identity, alias string, redirect map[string]routing.Redirect) error {
	var err error
	success := 0

	for _, v := range a.TrackerList {
		newErr := (&tracker.Router{
			Origin: a.Origin,
			URL:    v,
		}).Register(key, alias, redirect)
		if newErr == nil {
			success++
		} else {
			err = newErr
		}
	}

	if success == 0 && err != nil {
		return errors.New("All registration failed. Last Error: " + err.Error())
	} else {
		return nil
	}
}
