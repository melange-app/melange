package people

import (
	"fmt"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"
	"getmelange.com/backend/models"

	gdb "github.com/huntaub/go-db"
)

// List Management

type AddList struct {
	Tables map[string]gdb.Table
	Store  *models.Store
}

func (c *AddList) Handle(req *router.Request) framework.View {
	return framework.Error500
}

// RemoveList allows a user to remove a list from the database.
type RemoveList struct{}

// Post starts the removal.
func (c *RemoveList) Post(req *router.Request) framework.View {
	out := &models.List{}
	err := req.JSON(out)
	if err != nil {
		fmt.Println("Unable to deserialize list removal", err)
		return framework.Error500
	}

	_, err = req.Environment.Tables.List.Delete(out).
		Exec(req.Environment.Store)
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

// GetLists returns a list of the lists that users are using to sort
// their contacts.
type GetLists struct {
	Tables map[string]gdb.Table
	Store  *models.Store
}

// Get will fetch the list of lists.
func (c *GetLists) Handle(req *router.Request) framework.View {
	out := make([]*models.List, 0)
	err := req.Environment.Tables.List.Get().
		All(req.Environment.Store, &out)
	if err != nil {
		fmt.Println("Error getting contacts", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: out,
	}
}
