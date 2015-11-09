package people

import (
	"fmt"

	"getmelange.com/backend/models/identity"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"
	"getmelange.com/backend/models"
	gdb "github.com/huntaub/go-db"
)

type addContactRequest struct {
	Address string `json:"address"`
	Follow  bool   `json:"follow"`
	List    []int  `json:"list"`
}

// AddContact will create a new contact in the database.
type AddContact struct{}

// Post will execute the route.
func (c *AddContact) Post(req *router.Request) framework.View {
	out := &addContactRequest{}
	err := req.JSON(out)
	if err != nil {
		fmt.Println("Unable to decode body", err)
		return framework.Error500
	}

	contact := &models.Contact{}
	_, err = req.Environment.Tables.Contact.Insert(contact).Exec(req.Environment.Store)
	if err != nil {
		fmt.Println("Unable to insert contact", err)
		return framework.Error500
	}

	for _, v := range out.List {
		_, err = req.Environment.Tables.ContactMembership.Insert(
			&models.ContactMembership{
				Contact: &gdb.HasOne{
					Value: int(contact.Id),
				},
				List: &gdb.HasOne{
					Value: v,
				},
			}).Exec(req.Environment.Store)
		if err != nil {
			fmt.Println("Unable to insert contact in list", err)
		}
	}

	modelAddress := &identity.Address{
		Alias:      out.Address,
		Subscribed: out.Follow,
		Contact: &gdb.HasOne{
			Value: int(contact.Id),
		},
	}

	_, err = req.Environment.Tables.Address.Insert(modelAddress).
		Exec(req.Environment.Store)
	if err != nil {
		fmt.Println("Unable to insert address", err)
		return framework.Error500
	}

	contact.Identities = []*identity.Address{modelAddress}

	profile, err := req.Environment.Manager.GetProfile(out.Address)
	if err != nil {
		fmt.Println("Unable to retrieve profile for", out.Address)
		fmt.Println(err)

		return framework.Error500
	}

	contact.Profile = profile

	return &framework.JSONView{
		Content: contact,
	}
}

// RemoveContact will remove a contact from the database.
type RemoveContact struct{}

// Post will execute the removal.
func (r *RemoveContact) Post(req *router.Request) framework.View {
	out := &models.Contact{}
	err := req.JSON(out)
	if err != nil {
		fmt.Println("Unable to deserialize contact removal", err)
		return framework.Error500
	}

	_, err = req.Environment.Tables.Contact.Delete(out).
		Exec(req.Environment.Store)
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
type ListContacts struct{}

// Get will retrieve the list of contacts.
func (c *ListContacts) Get(req *router.Request) framework.View {
	out := make([]*models.Contact, 0)
	err := req.Environment.Tables.Contact.Get().All(req.Environment.Store, &out)
	if err != nil {
		fmt.Println("Error getting contacts", err)
		return framework.Error500
	}

	for _, v := range out {
		err := v.LoadIdentities(req.Environment.Store, req.Environment.Tables)
		if err != nil {
			fmt.Println("Couldn't get address, got error", err)
			continue
		}

		// Only attempt to populate the information if we have
		// at least one identity associated with a contact.
		if len(v.Identities) > 0 {
			profile, err := req.Environment.Manager.GetProfile(v.Identities[0].Alias)
			if err != nil {
				fmt.Println("Couldn't get profile, got error", err)
				continue
			}

			v.Profile = profile
		}
	}

	return &framework.JSONView{
		Content: out,
	}
}

type updateContactRequest struct {
	Id         int                 `json:"id"`
	Name       string              `json:"name"`
	Starred    bool                `json:"favorite"`
	Subscribed bool                `json:"subscribed"`
	Addresses  []*identity.Address `json:"addresses"`
}

// UpdateContact will update a contact in the database to a new value.
type UpdateContact struct {
	Tables map[string]gdb.Table
	Store  *models.Store
}

// Post will execute the update.
func (c *UpdateContact) Post(req *router.Request) framework.View {
	cr := &updateContactRequest{}
	err := req.JSON(cr)
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

	contactTable := req.Environment.Tables.Contact

	// Either update or insert the contact into the database.
	var query gdb.Statement
	if cr.Id < 0 {
		query = contactTable.Insert(contact)
	} else {
		query = contactTable.Update(contact)
	}

	// Execute the query.
	_, err = query.Exec(req.Environment.Store)
	if err != nil {
		fmt.Println("Error inserting contact", err)
		return framework.Error500
	}

	// Range over the provided addresses and store them in the
	// database.
	for _, v := range cr.Addresses {
		v.Contact = gdb.ForeignKey(contact)
		v.Subscribed = cr.Subscribed

		var query gdb.Statement
		if v.Id == 0 {
			query = req.Environment.Tables.Address.Insert(v)
		} else {
			query = req.Environment.Tables.Address.Update(v)
		}

		_, err = query.Exec(req.Environment.Store)
		if err != nil {
			fmt.Println("Error updating address", err)
			return framework.Error500
		}
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "",
	}
}
