package mailserver

// This file describes two routes that are used throughout the Melange System.

import (
	"airdispat.ch/identity"
	"airdispat.ch/routing"
	"airdispat.ch/tracker"
	"errors"
	"strings"
	"sync"
)

var cache map[string]*identity.Address
var cLock sync.RWMutex

var ServerKey *identity.Identity

func InitRouter() {
	if cache == nil {
		cache = make(map[string]*identity.Address)
	}
	if ServerKey == nil {
		ServerKey, _ = identity.CreateIdentity()
	}
	if RegistrationRouter == nil {
		RegistrationRouter = &tracker.TrackerRouter{
			tracker.GetTrackingServerLocationFromURL("airdispat.ch"),
			ServerKey,
		}
	}
	if LookupRouter == nil {
		LookupRouter = &Router{
			Origin:      ServerKey,
			TrackerList: []string{"mailserver.airdispat.ch:5000"},
		}
	}
}

var RegistrationRouter routing.Router
var LookupRouter routing.Router

type Router struct {
	Origin      *identity.Identity
	TrackerList []string
}

func (a *Router) LookupAlias(from string) (*identity.Address, error) {
	cLock.RLock()
	test, ok := cache[from]
	if ok {
		return test, nil
	}
	cLock.RUnlock()

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
		cLock.Lock()
		cache[from] = addr
		cache[addr.String()] = addr
		cLock.Unlock()
		return addr, nil
	}
	return nil, err
}

func (a *Router) Lookup(from string) (*identity.Address, error) {
	cLock.RLock()
	test, ok := cache[from]
	if ok {
		return test, nil
	}
	cLock.RUnlock()

	for _, v := range a.TrackerList {
		a, err := (&tracker.TrackerRouter{v, ServerKey}).Lookup(from)
		if err == nil {
			cLock.Lock()
			cache[from] = a
			cache[a.String()] = a
			cLock.Unlock()
			return a, nil
		}
	}
	return nil, errors.New("Couldn't find address in Trackers.")
}

func (a *Router) Register(key *identity.Identity, alias string) error {
	return errors.New("Can't use LookupRouter for registration.")
}
