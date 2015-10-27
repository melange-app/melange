package models

import (
	"fmt"

	"getmelange.com/backend/models/identity"
	"getmelange.com/backend/models/messages"
	gdb "github.com/huntaub/go-db"
)

func mapErrors(err ...error) error {
	for _, v := range err {
		if v != nil {
			return v
		}
	}
	return nil
}

func CreateTables(conn gdb.Database) (map[string]gdb.Table, error) {
	tables := make(map[string]gdb.Table)
	var err1, err2, err3, err4, err5, err6, err7, errList, errMembership error

	// identity.go
	tables["identity"], err1 = gdb.CreateTableFromStruct("identity", conn, false, &identity.Identity{})
	if err1 != nil {
		fmt.Println("err1", err1)
	}
	tables["alias"], err2 = gdb.CreateTableFromStruct("alias", conn, false, &identity.Alias{})
	if err2 != nil {
		fmt.Println("err2", err2)
	}
	tables["profile"], err7 = gdb.CreateTableFromStruct("profile", conn, false, &identity.Profile{})
	if err7 != nil {
		fmt.Println("err7", err2)
	}

	// contact.go
	tables["contact"], err3 = gdb.CreateTableFromStruct("contact", conn, false, &Contact{})
	if err3 != nil {
		fmt.Println("err3", err3)
	}
	tables["address"], err4 = gdb.CreateTableFromStruct("address", conn, false, &identity.Address{})
	if err4 != nil {
		fmt.Println("err4", err4)
	}

	// Lists (in contact.go)
	tables["list"], errList = gdb.CreateTableFromStruct("list", conn, false, &List{})
	if errList != nil {
		fmt.Println("errList", errList)
	}
	tables["contact_membership"], errMembership = gdb.CreateTableFromStruct("contact_membership", conn, false, &ContactMembership{})
	if errMembership != nil {
		fmt.Println("errMembership", errList)
	}

	// message.go
	tables["message"], err5 = gdb.CreateTableFromStruct("message", conn, false, &messages.Message{})
	if err5 != nil {
		fmt.Println("err5", err5)
	}
	tables["component"], err6 = gdb.CreateTableFromStruct("component", conn, false, &messages.Component{})
	if err6 != nil {
		fmt.Println("err6", err6)
	}

	return tables, mapErrors(err1, err2, err3, err4, err5, err6, err7, errList, errMembership)
}
