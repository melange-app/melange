package identity

import (
	"fmt"

	"getmelange.com/backend/models/identity"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"

	gdb "github.com/huntaub/go-db"
)

// RemoveIdentity will take an identity and remove it from the
// database. Note: this is super dangerous as there is no way to
// recover a deleted identity.
type RemoveIdentity struct{}

// Post will take an identity
func (r *RemoveIdentity) Post(req *router.Request) framework.View {
	request := &identity.Identity{}
	err := req.JSON(request)
	if err != nil {
		fmt.Println("Cannot decode body", err)
		return framework.Error500
	}

	if request.Fingerprint == "*" {
		return &framework.HTTPError{
			ErrorCode: 400,
			Message:   "Bad request: Cannot remove all identities.",
		}
	}

	// Delete where the fingerprint matches the one specified.
	_, err = (&gdb.DeleteStatement{
		Table: req.Environment.Tables.Identity.(*gdb.BasicTable).TableName,
		Where: &gdb.NamedEquality{
			Name:  "fingerprint",
			Value: request.Fingerprint,
		},
	}).Exec(req.Environment.Store)
	if err != nil {
		fmt.Println("Cannot remove identity.", err)
		return framework.Error500
	}

	fmt.Println("Removed the identity.", request.Fingerprint)
	return &framework.JSONView{
		Content: map[string]interface{}{
			"error": false,
		},
	}
}
