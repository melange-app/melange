package tracker

import (
	gdb "github.com/huntaub/go-db"
	"github.com/jmoiron/sqlx"
)

func CreateTables(conn *sqlx.DB) (gdb.Table, error) {
	return gdb.CreateTableFromStruct("record", conn, false, &Record{})
}

type Record struct {
	Id      gdb.PrimaryKey
	Address string
	Alias   string
	Message []byte
}
