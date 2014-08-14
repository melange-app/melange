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
	Tables map[string]gdb.Table
	Store  *models.Store
}

func (c *ListContacts) Handle(req *http.Request) framework.View {
	out := make([]*models.Contact, 0)
	err := c.Tables["contact"].Get().All(c.Store, &out)
	if err != nil {
		fmt.Println("Error getting contacts", err)
		return framework.Error500
	}

	for _, v := range out {
		v.Identities = make([]*models.Address, 0)
		err := c.Tables["address"].Get().Where("contact", v.Id).All(c.Store, &v.Identities)
		if err != nil {
			fmt.Println("Couldn't get address, got error", err)
		}
	}

	return &framework.JSONView{
		Content: out,
	}
}

type contactRequest struct {
	Id         int               `json:"id"`
	Name       string            `json: "name"`
	Starred    bool              `json:"favorite"`
	Subscribed bool              `json:"subscribed"`
	Addresses  []*models.Address `json:"addresses"`
}

type UpdateContact struct {
	Tables map[string]gdb.Table
	Store  *models.Store
}

func (c *UpdateContact) Handle(req *http.Request) framework.View {
	cr := &contactRequest{}
	err := DecodeJSONBody(req, &cr)
	if err != nil {
		fmt.Println("Error decoding JSON body", err)
		return framework.Error500
	}

	// Assemble contact
	contact := &models.Contact{
		Id:     gdb.PrimaryKey(cr.Id),
		Name:   cr.Name,
		Notify: cr.Starred,
	}

	if cr.Id < 0 {
		_, err = c.Tables["contact"].Insert(contact).Exec(c.Store)
		if err != nil {
			fmt.Println("Error inserting contact", err)
			return framework.Error500
		}
	} else {
		_, err = c.Tables["contact"].Update(contact).Exec(c.Store)
		if err != nil {
			fmt.Println("Can't update contact", err)
		}
	}

	for _, v := range cr.Addresses {
		v.Contact = gdb.ForeignKey(contact)
		v.Subscribed = cr.Subscribed

		if v.Id == 0 {
			_, err = c.Tables["address"].Insert(v).Exec(c.Store)
			if err != nil {
				fmt.Println("Error inserting address", err)
				return framework.Error500
			}
		} else {
			_, err = c.Tables["address"].Update(v).Exec(c.Store)
			if err != nil {
				fmt.Println("Error updating address", err)
				return framework.Error500
			}
		}
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "Done!",
	}
}
