package models

import (
	"getmelange.com/backend/models/db"
	"getmelange.com/backend/models/identity"
	"getmelange.com/backend/models/messages"
	gdb "github.com/huntaub/go-db"
)

type List struct {
	Id       gdb.PrimaryKey `json:"id"`
	Name     string         `json:"name"`
	Parent   *gdb.HasOne    `table:"list" json:"-"`
	Children *gdb.HasMany   `table:"list" on:"parent" json:"-"`
	Contacts *gdb.HasMany   `table:"contact_membership" on:"list" json:"-"`
}

type ContactMembership struct {
	Id      gdb.PrimaryKey `json:"id"`
	Contact *gdb.HasOne    `table:"contact" json:"-"`
	List    *gdb.HasOne    `table:"contact" json:"-"`
}

type Contact struct {
	Id        gdb.PrimaryKey `json:"id"`
	Name      string         `json:"name"`
	Image     string         `json:"image"`
	Notify    bool           `json:"favorite"`
	Addresses *gdb.HasMany   `table:"address" on:"contact" json:"-"`
	List      *gdb.HasMany   `table:"contact_membership" on:"contact" json:"-"`

	// JSON Specific Transient Fields
	Identities []*identity.Address   `db:"-" json:"addresses"`
	Profile    *messages.JSONProfile `db:"-" json:"profile"`
	Lists      []*List               `db:"-" json:"lists"`
}

func (c *Contact) LoadIdentities(store *Store, tables *db.Tables) error {
	c.Identities = make([]*identity.Address, 0)
	return tables.Address.Get().Where("contact", c.Id).All(store, &c.Identities)
}
