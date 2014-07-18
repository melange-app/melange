package app

import (
	"airdispat.ch/identity"
)

type Provider struct {
	Id          string
	Location    string
	Image       string
	Suffix      string
	Key         *identity.Address
	Description string
	Name        string
}

func GetServers() []*Provider {
	return []*Provider{
		&Provider{
			Id:          "1",
			Image:       "http://placehold.it/400",
			Location:    "airdispatch.me:1024",
			Key:         nil,
			Description: "The first AirDispatch provider. Dedicated to Melange.",
			Name:        "AirDispatch.Me",
		},
	}
}

func ServerKeyFromId(id string) *identity.Address {
	return nil
}

func GetTrackers() []*Provider {
	return []*Provider{
		&Provider{
			Id:          "1",
			Image:       "http://placehold.it/400",
			Location:    "airdispatch.me:2048",
			Suffix:      "airdispat.ch",
			Key:         nil,
			Description: "The first AirDispatch provider. Dedicated to Melange.",
			Name:        "AirDispatch.Me",
		},
	}
}

func TrackerKeyFromId(id string) *identity.Address {
	return nil
}
