package models

import (
	"airdispat.ch/identity"
	"airdispat.ch/routing"
	gdb "github.com/huntaub/go-db"
)


type Contact struct {
	Id         gdb.PrimaryKey `json:"id"`
	Name       string         `json:"name"`
	Image      string         `json:"image"`
	Notify     bool           `json:"favorite"`
	Addresses  *gdb.HasMany   `table:"address" on:"contact" json:"-"`
	Identities []*Address     `db:"-" json:"addresses"`
}

func (c *Contact) LoadProfile(r routing.Router, from *identity.Identity) error {
	currentAddress := c.Identities[0]

	fp := currentAddress.Fingerprint
	if fp == "" {
		temp, err := r.LookupAlias(currentAddress.Alias, routing.LookupTypeMAIL)
		if err != nil {
			return err
		}
		fp = temp.String()

	}

	profile, err := translateProfile(r, from, fp, currentAddress.Alias)
	if err != nil {
		return err
	}

	c.Profile = profile
	return nil
}

type Address struct {
	Id            gdb.PrimaryKey `json:"id"`
	Contact       *gdb.HasOne    `table:"contact" json:"-"`
	Alias         string         `json:"alias"`
	Fingerprint   string         `json:"fingerprint"`
	EncryptionKey []byte         `json:"-"`
	Location      string         `json:"location"`
	Subscribed    bool           `json:"subscribed"`
}
