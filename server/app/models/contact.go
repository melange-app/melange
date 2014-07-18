package models

import ()

type Contact struct {
	Name   string
	Image  string
	Notify bool
}

type Address struct {
	ContactId     int
	Fingerprint   string
	EncryptionKey []byte
}
