package messages

import (
	"fmt"
	"time"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"
	"getmelange.com/backend/models/identity"
	"getmelange.com/backend/models/messages"

	gdb "github.com/huntaub/go-db"
)

// CurrentProfile will return the current profile of the user.
type CurrentProfile struct{}

// Get will retrieve the profile.
func (c *CurrentProfile) Get(req *router.Request) framework.View {
	manager := req.Environment.Manager

	if !manager.HasIdentity {
		return &framework.HTTPError{
			ErrorCode: 422,
			Message:   "No profile to load yet.",
		}
	}

	profile := &identity.Profile{}
	err := manager.Identity.Profile.One(req.Environment.Store, profile)
	if err != nil {
		// This could occur if the user has just setup the
		// identity without providing a profile.
		fmt.Println("Can't get user's profile.", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: profile,
	}
}

// UpdateProfile allows a user to update their profile.
type UpdateProfile struct{}

// Post will actually perform the profile updating.
func (c *UpdateProfile) Post(req *router.Request) framework.View {
	profileRequest := &identity.Profile{}
	err := req.JSON(&profileRequest)
	if err != nil {
		fmt.Println("(Profile) Error occured while decoding update profile:", err)
		return framework.Error500
	}

	msg := &messages.JSONMessage{
		Name: "profile",
		Date: time.Now(),
		From: &messages.JSONProfile{
			Fingerprint: "",
			Alias:       "",
		},
		Public: true,
		Components: map[string]*messages.JSONComponent{
			"airdispat.ch/profile/name": &messages.JSONComponent{
				String: profileRequest.Name,
			},
			"airdispat.ch/profile/avatar": &messages.JSONComponent{
				String: string(profileRequest.Image),
			},
			"airdispat.ch/profile/description": &messages.JSONComponent{
				String: profileRequest.Description,
			},
		},
	}

	manager := req.Environment.Manager

	if manager.Identity.Profile.Value == 0 {
		// First Time Profile
		err = manager.PublishMessage(msg)
		if err != nil {
			fmt.Println("(Profile) Error publishing message", err)
			return framework.Error500
		}

		_, err = req.Environment.Tables.Profile.Insert(profileRequest).
			Exec(req.Environment.Store)
		if err != nil {
			fmt.Println("Couldn't put profile in database.")
			return framework.Error500
		}

		manager.Identity.Profile = gdb.ForeignKey(profileRequest)
		_, err = req.Environment.Tables.Identity.Update(manager.Identity).
			Exec(req.Environment.Store)
		if err != nil {
			fmt.Println("Couldn't update id with profile", err)
			return framework.Error500
		}

	} else {
		// Update Profile
		err := manager.UpdateMessage(msg)
		if err != nil {
			fmt.Println("(Profile) Error updating message", err)
			return framework.Error500
		}

		profileRequest.Id = gdb.PrimaryKey(manager.Identity.Profile.Value)
		_, err = req.Environment.Tables.Profile.Update(profileRequest).
			Exec(req.Environment.Store)
		if err != nil {
			fmt.Println("Couldn't update profile in db", err)
			return framework.Error500
		}

	}
	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "Successfully updated profile.",
	}
}
