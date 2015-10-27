package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/info"
	"getmelange.com/backend/models"
	"getmelange.com/backend/models/messages"
	"getmelange.com/backend/packaging"
	"getmelange.com/dap"
	"getmelange.com/router"
)

// Request is an HTTP request with the important fields already parsed
// out and ready for use.
type Request struct {
	Environment *info.Environment

	*http.Request
}

// JSON will parse the request body and return the deserialized JSON
// object.
func (r *Request) JSON(obj interface{}) error {
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()
	return decoder.Decode(obj)
}

// CurrentIdentityOrError will retrieve the current Identity object from the store.
func (r *Request) Identity() (*models.Identity, *framework.HTTPError) {
	store := r.Environment.Settings
	table := r.Environment.Tables()["identity"]

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
func (r *Request) Client() (*dap.Client, *framework.HTTPError) {
	store := r.Environment.Settings
	table := r.Environment.Tables()["identity"]

	id, err := CurrentIdentityOrError(store, table)
	if err != nil {
		return nil, err
	}

	cli, regErr := id.Client(store)
	if regErr != nil {
		fmt.Println("Error getting DAPClientFromID", err)
		return nil, framework.Error500
	}

	return cli, nil
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

func parseRequest(req *http.Request, env *info.Environment) *Request {
	return &Request{
		Request:     req,
		Environment: env,
	}
}
