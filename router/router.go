package router

// This file describes two routes that are used throughout the Melange System.

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"airdispat.ch/identity"
	"airdispat.ch/routing"
	"airdispat.ch/tracker"
	cache "github.com/huntaub/go-cache"
)

var routeCache *cache.Cache = cache.NewCache(1 * time.Hour)
var knownTrackers []string = []string{"localhost:2048"}

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
	b := &Router{
		Origin:      a.Origin,
		TrackerList: a.TrackerList,
	}
	return b.lookupAliasHelper(from, name)
}

func (a *Router) lookupAliasHelper(from string, name routing.LookupType) (*identity.Address, error) {
	key := fmt.Sprintf("%s:%s", from, name)
	test, stale := routeCache.Get(key)
	if !stale {
		return test.(*identity.Address), nil
	}

	if from == "" {
		return nil, errors.New("Can't lookup nothing.")
	}

	if from[0] == '/' {
		return a.Lookup(from[1:], name)
	}

	comp := strings.Split(from, "@")

	if len(comp) != 2 {
		return nil, errors.New("Can't use lookup router without tracker address.")
	}

	url := getTrackerURL(comp[1])

	t := &tracker.Router{
		URL:        url,
		Origin:     a.Origin,
		Redirector: a,
	}

	addr, err := t.LookupAlias(comp[0], name)
	if err == nil {
		routeCache.Store(key, addr)
		routeCache.Store(fmt.Sprintf("%s:%s", addr, name), addr)
		return addr, nil
	}
	return nil, err
}

func (a *Router) Lookup(from string, name routing.LookupType) (*identity.Address, error) {
	b := &Router{
		Origin:      a.Origin,
		TrackerList: a.TrackerList,
	}
	return b.lookupHelper(from, name)
}

func (a *Router) lookupHelper(from string, name routing.LookupType) (*identity.Address, error) {
	key := fmt.Sprintf("%s:%s", from, name)
	test, stale := routeCache.Get(key)
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
			routeCache.Store(key, a)
			routeCache.Store(fmt.Sprintf("%s:%s", a, name), a)
			return a, nil
		}
	}
	return nil, errors.New("Couldn't find address in Trackers.")
}

func (a *Router) Register(key *identity.Identity, alias string, redirect map[string]routing.Redirect) error {
	var err error
	success := 0

	for _, v := range a.TrackerList {
		url := getTrackerURL(v)

		newErr := (&tracker.Router{
			Origin: a.Origin,
			URL:    url,
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
