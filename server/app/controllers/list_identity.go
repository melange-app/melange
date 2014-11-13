package controllers

import (
	"fmt"
	"net/http"

	"getmelange.com/app/framework"
	"getmelange.com/app/models"

	gdb "github.com/huntaub/go-db"
)

type RemoveIdentity struct {
	Tables map[string]gdb.Table
	Store *models.Store
}

func (r *RemoveIdentity) Handle(req *http.Request) framework.View {
	request := &models.Identity{}
	err := DecodeJSONBody(req, request)
	if err != nil {
		fmt.Println("Cannot decode body", err)
		return framework.Error500
	}

	_, err = (&gdb.DeleteStatement{
		Table: r.Tables["identity"].(*gdb.BasicTable).TableName,
		Where: &gdb.NamedEquality {
			Name: "fingerprint",
			Value: request.Fingerprint,
		},
	}).Exec(r.Store)
	if err != nil {
		fmt.Println("Cannot remove identity.", err)
		return framework.Error500
	}

	fmt.Println("Removed the identity.", request.Fingerprint)
	return &framework.JSONView {
		Content: map[string]interface{} {
			"error": false,
		},
	}
}

// ListIdentity will list all identities on file.
type ListIdentity struct {
	Tables map[string]gdb.Table
	Store  *models.Store
}

// Handle will get all identities then return a JSONView with them.
func (i *ListIdentity) Handle(req *http.Request) framework.View {
	results := make([]*models.Identity, 0)
	err := i.Tables["identity"].Get().All(i.Store, &results)
	if err != nil {
		fmt.Println("Error getting Identities:", err)
		return framework.Error500
	}

	fingerprint, err := i.Store.Get("current_identity")
	if err != nil {
		fmt.Println("Error getting current identity.", err)
		return framework.Error500
	}

	for _, v := range results {
		if v.Fingerprint == fingerprint {
			v.Current = true
		}
	}

	return &framework.JSONView{
		Content: results,
	}
}
