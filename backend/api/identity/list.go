package identity

import (
	"fmt"
	"net/http"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/models"
)

// ListIdentity will list all identities on file.
type ListIdentity struct{}

// Get will return a list of all the identities currently in the
// system for display in a settings-view.
func (i *ListIdentity) Get(req *http.Request) framework.View {
	results := make([]*models.Identity, 0)
	err := req.Environment.Tables()["identity"].Get().All(i.Store, &results)
	if err != nil {
		fmt.Println("Error getting Identities:", err)
		return framework.Error500
	}

	fingerprint, err := req.Environment.Settings.Get("current_identity")
	if err != nil {
		fmt.Println("Error getting current identity.", err)
		return framework.Error500
	}

	for _, v := range results {
		aliases := make([]*models.Alias, 0)
		err := i.Tables["alias"].Get().Where("identity", v.Id).All(req.Environment.Settings, &aliases)
		if err != nil {
			fmt.Println("Error loading identity aliases", err)
		}
		v.LoadedAliases = aliases

		if v.Fingerprint == fingerprint {
			v.Current = true
		}
	}

	return &framework.JSONView{
		Content: results,
	}
}
