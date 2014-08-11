package models

import gdb "github.com/huntaub/go-db"

type Contact struct {
	Id        gdb.PrimaryKey `json:"id"`
	Name      string         `json:"name"`
	Image     string         `json:"image"`
	Notify    bool           `json:"favorite"`
	Addresses *gdb.HasMany   `table:"address" on:"contact"`
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
