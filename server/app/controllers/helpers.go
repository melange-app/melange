package controllers

import (
	"encoding/json"
	"fmt"
	"melange/app/framework"
	"melange/app/models"
	"melange/dap"
	"net/http"

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

	adID, regErr := id.ToDispatch("")
	if regErr != nil {
		fmt.Println("Error serializing identity.", regErr)
		return nil, framework.Error500
	}

	server, regErr := id.CreateServerFromIdentity()
	if regErr != nil {
		fmt.Println("Error getting current server.", regErr)
		return nil, framework.Error500
	}

	return &dap.Client{
		Key:    adID,
		Server: server,
	}, nil
}
