package people

import (
	"fmt"
	"net/http"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/models"
)

// List Management

type AddList struct {
	Tables map[string]gdb.Table
	Store  *models.Store
}

func (c *AddList) Handle(req *http.Request) framework.View {
	return framework.Error500
}

type RemoveList struct {
	Tables map[string]gdb.Table
	Store  *models.Store
}

func (c *RemoveList) Handle(req *http.Request) framework.View {
	out := &models.List{}
	err := DecodeJSONBody(req, out)
	if err != nil {
		fmt.Println("Unable to deserialize list removal", err)
		return framework.Error500
	}

	_, err = c.Tables["list"].Delete(out).Exec(c.Store)
	if err != nil {
		fmt.Println("Unable to delete list", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: map[string]interface{}{
			"error": false,
		},
	}
}

type GetLists struct {
	Tables map[string]gdb.Table
	Store  *models.Store
}

func (c *GetLists) Handle(req *http.Request) framework.View {
	out := make([]*models.List, 0)
	err := c.Tables["list"].Get().All(c.Store, &out)
	if err != nil {
		fmt.Println("Error getting contacts", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: out,
	}
}
