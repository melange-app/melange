package models

import (
	"github.com/coopernurse/gorp"
)

func CreateTables(g *gorp.DbMap) {
	t := g.AddTableWithName(User{}, "dispatch_user").SetKeys(true, "UserId")
	t.ColMap("Password").Transient = true
}
