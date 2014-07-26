package controllers

import (
	"fmt"
	"melange/app/framework"
	"melange/app/models"
	"net/http"

	gdb "github.com/huntaub/go-db"
)

// ListIdentity will list all identities on file.
type ListIdentity struct {
	Tables map[string]gdb.Table
	Store  *models.Store
}

// Handle will get all identities then return a JSONView with them.
func (i *ListIdentity) Handle(req *http.Request) framework.View {
	var results []*models.Identity
	err := i.Tables["identity"].Get().All(i.Store, &results)
	if err != nil {
		fmt.Println("Error getting Identities:", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: results,
	}
}
