package controllers

import (
	"fmt"
	"net/http"

	"getmelange.com/app/framework"
	"getmelange.com/app/messages"
	"getmelange.com/app/models"
	"getmelange.com/router"

	"airdispat.ch/identity"
	gdb "github.com/huntaub/go-db"
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

// Contact Management

type AddContact struct {
	Tables map[string]gdb.Table
	Store  *models.Store
}

func (c *AddContact) Handle(req *http.Request) framework.View {
	out := make(map[string]interface{})
	err := DecodeJSONBody(req, &out)
	if err != nil {
		fmt.Println("Unable to decode body", err)
		return framework.Error500
	}

	tempAddress, ok1 := out["address"]
	address, ok2 := tempAddress.(string)
	if !(ok1 || ok2) {
		fmt.Println("Unable to decode address")
		return framework.Error500
	}

	tempFollow, ok1 := out["follow"]
	follow, ok2 := tempFollow.(bool)
	if !(ok1 || ok2) {
		fmt.Println("Unable to decode follow")
		return framework.Error500
	}

	tempList, ok1 := out["list"]
	list, ok2 := tempList.([]int)
	if !(ok1 || ok2) {
		fmt.Println("Unable to decode list")
		return framework.Error500
	}
	fmt.Println(address, follow, list)

	contact := &models.Contact{
	//		List: &gdb.HasOne{
	//			Value: list,
	//		},
	}
	_, err = c.Tables["contact"].Insert(contact).Exec(c.Store)
	if err != nil {
		fmt.Println("Unable to insert contact", err)
		return framework.Error500
	}

	for _, v := range list {
		_, err = c.Tables["contact_membership"].Insert(&models.ContactMembership{
			Contact: &gdb.HasOne{
				Value: int(contact.Id),
			},
			List: &gdb.HasOne{
				Value: v,
			},
		}).Exec(c.Store)
	}

	modelAddress := &models.Address{
		Alias:      address,
		Subscribed: follow,
		Contact: &gdb.HasOne{
			Value: int(contact.Id),
		},
	}
	_, err = c.Tables["address"].Insert(modelAddress).Exec(c.Store)
	if err != nil {
		fmt.Println("Unable to insert address", err)
		return framework.Error500
	}

	contact.Identities = []*models.Address{modelAddress}

	retriever, frameErr := createContactsProfileRetriever(c.Store, c.Tables)
	if frameErr != nil {
		return frameErr
	}

	err = messages.LoadContactProfile(retriever.router, contact, retriever.id)
	if err != nil {
		fmt.Println("Unable to load profile", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: contact,
	}
}

type RemoveContact struct {
	Tables map[string]gdb.Table
	Store  *models.Store
}

func (r *RemoveContact) Handle(req *http.Request) framework.View {
	out := &models.Contact{}
	err := DecodeJSONBody(req, out)
	if err != nil {
		fmt.Println("Unable to deserialize contact removal", err)
		return framework.Error500
	}

	_, err = r.Tables["contact"].Delete(out).Exec(r.Store)
	if err != nil {
		fmt.Println("Unable to delete contact", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: map[string]interface{}{
			"error": false,
		},
	}
}

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

	// Getting Profile Information
	retriever, frameErr := createContactsProfileRetriever(c.Store, c.Tables)
	if frameErr != nil {
		return frameErr
	}

	for _, v := range out {
		err := v.LoadIdentities(c.Store, c.Tables)
		if err != nil {
			fmt.Println("Couldn't get address, got error", err)
			continue
		}

		err = messages.LoadContactProfile(retriever.router, v, retriever.id)
		if err != nil {
			fmt.Println("Couldn't get profile, got error", err)
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
