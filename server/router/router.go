package router

// This file describes two routes that are used throughout the Melange System.

import (
	"airdispat.ch/identity"
	"airdispat.ch/tracker"
	"errors"
	"fmt"
	"github.com/huntaub/go-cache"
	"strings"
)

var routeCache *cache.Cache

// var cLock sync.RWMutex

var ServerKey *identity.Identity

type Router struct {
	Origin      *identity.Identity
	TrackerList []string
}

func (a *Router) LookupAlias(from string) (*identity.Address, error) {
	test, stale := routeCache.Get(from)
	if !stale {
		return test.(*identity.Address), nil
	}

	if from[0] == '/' {
		return a.Lookup(from[1:])
	}
	comp := strings.Split(from, "@")
	if len(comp) != 2 {
		return nil, errors.New("Can't use lookup router without tracker address.")
	}

	url := tracker.GetTrackingServerLocationFromURL(comp[1])
	t := &tracker.TrackerRouter{url, a.Origin}

	addr, err := t.LookupAlias(comp[0])
	if err == nil {
		routeCache.Store(from, addr)
		routeCache.Store(addr.String(), addr)
		return addr, nil
	}
	return nil, err
}

func (a *Router) Lookup(from string) (*identity.Address, error) {
	test, stale := routeCache.Get(from)
	if !stale {
		return test.(*identity.Address), nil
	}

	for _, v := range a.TrackerList {
		fmt.Println(v, ServerKey)
		a, err := (&tracker.TrackerRouter{v, ServerKey}).Lookup(from)
		if err == nil {
			routeCache.Store(from, a)
			routeCache.Store(a.String(), a)
			return a, nil
		}
	}
	return nil, errors.New("Couldn't find address in Trackers.")
}

func (a *Router) Register(key *identity.Identity, alias string) error {
	var err error
	success := 0

	for _, v := range a.TrackerList {
		newErr := (&tracker.TrackerRouter{
			Origin: a.Origin,
			URL:    v,
		}).Register(key, alias)
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
