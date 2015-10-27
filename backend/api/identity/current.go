package identity

import (
	"fmt"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"
	"getmelange.com/backend/info"
	"getmelange.com/backend/models/messages"
)

type setCurrentIdentity struct {
	Fingerprint string `json:"fingerprint"`
}

// CurrentIdentity manages a REST-ful service for updating and retrieving the
// currently utilized identity.
type CurrentIdentity struct{}

// Post will take the Fingerprint supplied in the Request body and set
// the current identity to the corresponding identity.
func (i *CurrentIdentity) Post(req *router.Request, env *info.Environment) framework.View {
	setRequest := &setCurrentIdentity{}
	err := req.JSON(&setRequest)
	if err != nil {
		fmt.Println("Cannot decode body", err)
		return framework.Error500
	}

	err = env.Settings.Set("current_identity", setRequest.Fingerprint)
	if err != nil {
		fmt.Println("Error storing current_identity.", err)
		return framework.Error500
	}

	messages.InvalidateCaches()

	return &framework.JSONView{
		Content: map[string]interface{}{
			"error": false,
		},
	}
}

// Get will return the current identity.
func (i *CurrentIdentity) Get(req *router.Request, env *info.Environment) framework.View {
	idTable := env.GetTables()["identity"]

	result, err := CurrentIdentityOrError(env.Settings, idTable)
	if err != nil {
		return err
	}

	return &framework.JSONView{
		Content: result,
	}
}

// Compile time checks that we adhere to the router interfaces.
var _ router.GetHandler = &CurrentIdentity{}
var _ router.PostHandler = &CurrentIdentity{}
