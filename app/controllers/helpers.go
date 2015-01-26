package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"getmelange.com/app/framework"
	"getmelange.com/app/messages"
	"getmelange.com/app/models"
	"getmelange.com/app/packaging"
	"getmelange.com/dap"
	"getmelange.com/router"

	gdb "github.com/huntaub/go-db"
)

// DecodeJSONBody will take a request and a destination interface and
// decode the JSON from the request into object.
func DecodeJSONBody(req *http.Request, object interface{}) error {
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()
	return decoder.Decode(object)
}

// CurrentIdentityOrError will retrieve the current Identity object from the store.
func CurrentIdentityOrError(store *models.Store, table gdb.Table) (*models.Identity, *framework.HTTPError) {
	data, err := store.GetDefault("current_identity", "")
	if err != nil {
		fmt.Println("Error getting current identity.", err)
		return nil, framework.Error500
	}

	if data == "" {
		return nil, &framework.HTTPError{
			ErrorCode: 422,
			Message:   "Cannot fufill current identity request before creating an identity.",
		}
	}

	result := &models.Identity{}
	err = table.Get().Where("fingerprint", data).One(store, result)
	if err != nil {
		fmt.Println("Error getting identity:", err)
		return nil, framework.Error500
	}

	return result, nil
}

// CurrentDAPClient will return the current DAPClient associated with the curernt
// Identity.
func CurrentDAPClient(store *models.Store, table gdb.Table) (*dap.Client, *framework.HTTPError) {
	id, err := CurrentIdentityOrError(store, table)
	if err != nil {
		return nil, err
	}

	cli, regErr := DAPClientFromID(id, store)
	if regErr != nil {
		fmt.Println("Error getting DAPClientFromID", err)
		return nil, framework.Error500
	}

	return cli, nil
}

func DAPClientFromID(id *models.Identity, store *models.Store) (*dap.Client, error) {
	adID, regErr := id.ToDispatch(store, "")
	if regErr != nil {
		return nil, regErr
	}

	server, regErr := id.CreateServerFromIdentity()
	if regErr != nil {
		return nil, regErr
	}

	return &dap.Client{
		Key:    adID,
		Server: server,
	}, nil
}

// ConstructManager will use the Store and Tables to create a *models.MessageManager.
func constructManager(
	store *models.Store,
	tables map[string]gdb.Table,
	p *packaging.Packager,
) (*messages.MessageManager, error) {
	id, err := CurrentIdentityOrError(store, tables["identity"])
	if err != nil {
		return nil, errors.New("Couldn't get current Identity.")
	}

	client, realErr := DAPClientFromID(id, store)
	if realErr != nil {
		return nil, realErr
	}

	router := &router.Router{
		Origin: client.Key,
		TrackerList: []string{
			"localhost:2048",
		},
	}

	return &messages.MessageManager{
		Tables:   tables,
		Store:    store,
		Client:   client,
		Identity: id,
		Router:   router,
		Packager: p,
	}, nil

}
