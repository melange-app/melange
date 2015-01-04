package models

import (
	"time"

	gdb "github.com/huntaub/go-db"
)

type MigrationFunc func(gdb.Database) error

func (m MigrationFunc) Apply(g gdb.Database) error {
	return m(g)
}

type Migration interface {
	Apply(gdb.Database) error
}

type MigrationRecord struct {
	Id          gdb.PrimaryKey
	Migration   int
	DateApplied *time.Time
}
