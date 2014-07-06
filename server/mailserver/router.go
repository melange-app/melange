package mailserver

// This file describes two routes that are used throughout the Melange System.

import (
	"airdispat.ch/identity"
	"airdispat.ch/routing"
	"airdispat.ch/tracker"
	"errors"
	"fmt"
	"github.com/huntaub/go-cache"
	"strings"
	"time"
)

var routeCache *cache.Cache

// var cLock sync.RWMutex

var ServerKey *identity.Identity

func InitRouter() {
	if routeCache == nil {
		routeCache = cache.NewCache(1 * time.Hour)
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
	return errors.New("Can't use LookupRouter for registration.")
}
