package models

import (
	"github.com/huntaub/go-db"
	"fmt"
)

func mapErrors(err ...error) error {
	for _, v := range err {
		if v != nil {
			return v
		}
	}
	return nil
}

func CreateTables(conn db.Executor) (map[string]db.Table, error) {
	tables := make(map[string]db.Table)
	var err1, err2, err3, err4, err5, err6 error

	// identity.go
	tables["identity"], err1 = db.CreateTableFromStruct("identity", conn, false, &Identity{})
	if err1 != nil {
		fmt.Println("err1", err1)
	}
	tables["alias"], err2 = db.CreateTableFromStruct("alias", conn, false, &Alias{})
	if err2 != nil {
		fmt.Println("err2", err2)
	}

	// contact.go
	tables["contact"], err3 = db.CreateTableFromStruct("contact", conn, false, &Contact{})
	if err3 != nil {
		fmt.Println("err3", err3)
	}
	tables["address"], err4 = db.CreateTableFromStruct("address", conn, false, &Address{})
	if err4 != nil {
		fmt.Println("err4", err4)
	}

	// message.go
	tables["message"], err5 = db.CreateTableFromStruct("message", conn, false, &Message{})
	if err5 != nil {
		fmt.Println("err5", err5)
	}
	tables["component"], err6 = db.CreateTableFromStruct("component", conn, false, &Component{})
	if err6 != nil {
		fmt.Println("err6", err6)
	}

	return tables, mapErrors(err1, err2, err3, err4, err5, err6)
}
