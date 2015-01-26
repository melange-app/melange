package models

import gdb "github.com/huntaub/go-db"

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
	Identities []*Address   `db:"-" json:"addresses"`
	Profile    *JSONProfile `db:"-" json:"profile"`
	Lists      []*List      `db:"-" json:"lists"`
}

func (c *Contact) LoadIdentities(store *Store, tables map[string]gdb.Table) error {
	c.Identities = make([]*Address, 0)
	return tables["address"].Get().Where("contact", c.Id).All(store, &c.Identities)
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
