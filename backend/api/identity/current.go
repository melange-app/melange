package identity

import (
	"fmt"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"
	"getmelange.com/backend/models/identity"
)

type setCurrentIdentity struct {
	Fingerprint string `json:"fingerprint"`
}

// CurrentIdentity manages a REST-ful service for updating and retrieving the
// currently utilized identity.
type CurrentIdentity struct{}

// Post will take the Fingerprint supplied in the Request body and set
// the current identity to the corresponding identity.
func (i *CurrentIdentity) Post(req *router.Request) framework.View {
	setRequest := &setCurrentIdentity{}
	err := req.JSON(&setRequest)
	if err != nil {
		fmt.Println("Cannot decode body", err)
		return framework.Error500
	}

	// Get the identity out of the database.
	newIdentity := &identity.Identity{}
	err = req.Environment.Tables.Identity.Get().Where("fingerprint", setRequest.Fingerprint).
		One(req.Environment.Store, newIdentity)
	if err != nil {
		fmt.Println("Unable to get new current identity", err)
		return framework.Error500
	}

	// Update the identity that we are using to do fetching.
	err = req.Environment.Manager.SwitchIdentity(newIdentity)
	if err != nil {
		fmt.Println("Received error switching identity", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: map[string]interface{}{
			"error": false,
		},
	}
}

// Get will return the current identity.
func (i *CurrentIdentity) Get(req *router.Request) framework.View {
	return &framework.JSONView{
		Content: req.Environment.Manager.Identity,
	}
}

// Compile time checks that we adhere to the router interfaces.
var _ router.GetHandler = &CurrentIdentity{}
var _ router.PostHandler = &CurrentIdentity{}
