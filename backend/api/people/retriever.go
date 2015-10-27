package people

import (
	"fmt"

	"airdispat.ch/identity"
	"getmelange.com/backend/framework"
	"getmelange.com/backend/models"
	"getmelange.com/router"
)

type contactsProfileRetriever struct {
	modelId *models.Identity
	id      *identity.Identity
	router  *router.Router
}

func createContactsProfileRetriever(store *models.Store, tables map[string]gdb.Table) (*contactsProfileRetriever, framework.View) {
	modelId, frameErr := CurrentIdentityOrError(store, tables["identity"])
	if frameErr != nil {
		return nil, frameErr
	}

	id, err := modelId.ToDispatch(store, "")
	if err != nil {
		fmt.Println("Error converting identity", err)
		return nil, framework.Error500
	}

	router := &router.Router{
		Origin: id,
		TrackerList: []string{
			"localhost:2048",
		},
	}

	return &contactsProfileRetriever{
		modelId: modelId,
		id:      id,
		router:  router,
	}, nil
}
