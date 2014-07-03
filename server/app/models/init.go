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

	// Add User Identity Tables
	g.AddTableWithName(Identity{}, "dispatch_identity").SetKeys(true, "IdentityId")

	// Add MailServer Tables
	g.AddTableWithName(Message{}, "dispatch_messages").SetKeys(true, "MessageId")
	g.AddTableWithName(Alert{}, "dispatch_alerts").SetKeys(true, "AlertId")
	g.AddTableWithName(Component{}, "dispatch_components").SetKeys(true, "ComponentId")
}
