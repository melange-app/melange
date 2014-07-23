package models

import (
	"github.com/huntaub/go-db"
)

type Contact struct {
	Id        db.PrimaryKey
	Name      string
	Image     string
	Notify    bool
	Addresses *db.HasMany `table:"address" on:"contact"`
}

type Address struct {
	Id            db.PrimaryKey
	Contact       *db.HasOne `table:"contact"`
	Fingerprint   string
	EncryptionKey []byte
	Location      string
}
