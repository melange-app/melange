package models

import (
	"github.com/huntaub/go-db"
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
	tables["alias"], err2 = db.CreateTableFromStruct("alias", conn, false, &Alias{})

	// contact.go
	tables["contact"], err3 = db.CreateTableFromStruct("contact", conn, false, &Contact{})
	tables["address"], err4 = db.CreateTableFromStruct("address", conn, false, &Address{})

	// message.go
	tables["message"], err5 = db.CreateTableFromStruct("message", conn, false, &Message{})
	tables["component"], err6 = db.CreateTableFromStruct("identity", conn, false, &Component{})

	return tables, mapErrors(err1, err2, err3, err4, err5, err6)
}
