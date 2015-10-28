package identity

import (
	"fmt"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/models/identity"

	"getmelange.com/backend/framework"
)

// ListIdentity will list all identities on file.
type ListIdentity struct{}

// Get will return a list of all the identities currently in the
// system for display in a settings-view.
func (i *ListIdentity) Get(req *router.Request) framework.View {
	results := new([]*identity.Identity)
	err := req.Environment.Tables.Identity.Get().All(req.Environment.Store, &results)
	if err != nil {
		fmt.Println("Error getting Identities:", err)
		return framework.Error500
	}

	current := req.Environment.Manager.Identity

	// Load the aliases for each of the addresses.
	for _, v := range *results {
		aliases := new([]*identity.Alias)
		err := req.Environment.Tables.Alias.
			Get().Where("identity", v.Id).All(req.Environment.Store, &aliases)
		if err != nil {
			fmt.Println("Error loading identity aliases", err)
		}
		v.LoadedAliases = *aliases

		// Ensure that the current id is noted as current if
		// it is selected.
		if v.Id == current.Id {
			v.Current = true
		}
	}

	return &framework.JSONView{
		Content: *results,
	}
}
