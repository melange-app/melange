package models

import (
	"github.com/coopernurse/gorp"
)

func CreateTables(g *gorp.DbMap) {
	// Add User Table
	t := g.AddTableWithName(User{}, "dispatch_user").SetKeys(true, "UserId")
	t.ColMap("Password").Transient = true

	// Add UserApp Table
	g.AddTableWithName(UserApp{}, "dispatch_app").SetKeys(true, "UserAppId")

	// Add UserSubscription Table
	g.AddTableWithName(UserSubscription{}, "dispatch_subscription").SetKeys(true, "SubscriptionId")
}
