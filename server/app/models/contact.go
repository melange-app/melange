package models

import (
	gdb "github.com/huntaub/go-db"
)

type Contact struct {
	Id        gdb.PrimaryKey
	Name      string
	Image     string
	Notify    bool
	Addresses *gdb.HasMany `table:"address" on:"contact"`
}

type Address struct {
	Id            gdb.PrimaryKey
	Contact       *gdb.HasOne `table:"contact"`
	Fingerprint   string
	EncryptionKey []byte
	Location      string
	Subscribed    bool
}
