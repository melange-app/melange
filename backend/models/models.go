package models

import (
	"getmelange.com/backend/models/db"
	"getmelange.com/backend/models/identity"
	"getmelange.com/backend/models/messages"
	gdb "github.com/huntaub/go-db"
)

// CreateTables will take a gdb.Database and ensure that all of the
// tables that Melange needs are created and loaded.
func CreateTables(conn gdb.Database) (*db.Tables, error) {
	tables := &db.Tables{}

	return tables, db.CreateTables(conn,
		db.Create("identity", &tables.Identity, &identity.Identity{}),
		db.Create("alias", &tables.Alias, &identity.Alias{}),
		db.Create("profile", &tables.Profile, &identity.Profile{}),

		db.Create("contact", &tables.Contact, &Contact{}),
		db.Create("address", &tables.Address, &identity.Address{}),
		db.Create("list", &tables.List, &List{}),
		db.Create("contact_membership", &tables.ContactMembership, &ContactMembership{}),

		db.Create("message", &tables.Message, &messages.Message{}),
		db.Create("component", &tables.Component, &messages.Component{}),
	)
}
