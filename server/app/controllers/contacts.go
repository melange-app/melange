package controllers

import (
	"fmt"
	"melange/app/framework"
	"melange/app/models"
	"net/http"

	gdb "github.com/huntaub/go-db"
)

// ListContacts will return a list (in JSON) of all of the contacts stored
// for an identity.
type ListContacts struct {
	Table gdb.Table
	Store *models.Store
}

func (c *ListContacts) Handle(req *http.Request) framework.View {
	out := make([]*models.Contact, 0)
	err := c.Table.Get().All(c.Store, &out)
	if err != nil {
		fmt.Println("Error getting contacts", err)
		return framework.Error500
	}
	return &framework.JSONView{
		Content: out,
	}
}

type contactRequest struct {
	Name       string `json: "name"`
	Address    string `json:"address"`
	Alias      string `json:"alias"`
	Subscribed bool   `json:"subscribed"`
	Starred    bool   `json:"starred"`
}

type NewContact struct {
	Tables map[string]gdb.Table
	Store  *models.Store
}

func (c *NewContact) Handle(req *http.Request) framework.View {
	cr := &contactRequest{}
	err := DecodeJSONBody(req, cr)
	if err != nil {
		fmt.Println("Error decoding JSON body", err)
		return framework.Error500
	}

	contact := &models.Contact{
		Name:   cr.Name,
		Notify: cr.Starred,
	}
	_, err = c.Tables["contact"].Insert(contact).Exec(c.Store)
	if err != nil {
		fmt.Println("Error inserting contact", err)
		return framework.Error500
	}

	address := &models.Address{
		Fingerprint: cr.Address,
		Subscribed:  cr.Subscribed,
		Contact:     gdb.ForeignKey(contact),
	}
	_, err = c.Tables["address"].Insert(address).Exec(c.Store)
	if err != nil {
		fmt.Println("Error inserting address", err)
		return framework.Error500
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "Done!",
	}
}
