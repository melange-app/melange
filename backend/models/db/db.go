package db

import (
	"fmt"

	gdb "github.com/huntaub/go-db"
)

// Store is an object that can perform queries.
type Store gdb.Executor

// Tables is an object that holds references to each of the tables in
// the database.
type Tables struct {
	// models/identity
	Identity gdb.Table
	Alias    gdb.Table
	Profile  gdb.Table

	// models/contacts
	Contact           gdb.Table
	Address           gdb.Table
	List              gdb.Table
	ContactMembership gdb.Table

	// models/messages
	Message   gdb.Table
	Component gdb.Table

	errors map[string]error
}

type request struct {
	Name     string
	Location *gdb.Table
	Struct   interface{}
}

// Create will make a request to create a table in a location with a
// specific type of data.
func Create(name string, loc *gdb.Table, data interface{}) request {
	return request{
		Name:     name,
		Location: loc,
		Struct:   data,
	}
}

func CreateTables(conn gdb.Database, requests ...request) error {
	var err error

	for _, req := range requests {
		*req.Location, err = gdb.CreateTableFromStruct(req.Name, conn, false, req.Struct)
		if err != nil {
			fmt.Println("[TAB-INIT]", "Received error creating", req.Name, "table.")
			fmt.Println("[TAB-INIT]", err)
		}
	}

	if err == nil {
		fmt.Println("Successfully created", len(requests), "tables.")
	}

	return err
}
