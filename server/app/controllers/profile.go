package controllers

import (
	"fmt"
	"net/http"
	"time"

	"getmelange.com/app/framework"
	"getmelange.com/app/models"

	gdb "github.com/huntaub/go-db"
)

type CurrentProfile struct {
	Tables map[string]gdb.Table
	Store  *models.Store
}

func (c *CurrentProfile) Handle(req *http.Request) framework.View {
	result, frameErr := CurrentIdentityOrError(c.Store, c.Tables["identity"])
	if frameErr != nil {
		return frameErr
	}

	if result.Profile.Value == 0 {
		return &framework.HTTPError{
			ErrorCode: 422,
			Message:   "No profile to load yet.",
		}
	}

	profile := &models.Profile{}
	err := result.Profile.One(c.Store, profile)
	if err != nil {
		fmt.Println("Can't get user's profile.", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: profile,
	}
}

type UpdateProfile struct {
	Tables map[string]gdb.Table
	Store  *models.Store
}

func (c *UpdateProfile) Handle(req *http.Request) framework.View {
	profileRequest := &models.Profile{}
	err := DecodeJSONBody(req, &profileRequest)
	if err != nil {
		fmt.Println("(Profile) Error occured while decoding update profile:", err)
		return framework.Error500
	}

	result, frameErr := CurrentIdentityOrError(c.Store, c.Tables["identity"])
	if frameErr != nil {
		return frameErr
	}

	cli, err := DAPClientFromID(result, c.Store)
	if err != nil {
		fmt.Println("(Profile) Error getting DAP Client from ID", err)
		return framework.Error500
	}

	msg := &melangeMessage{
		Name: "profile",
		Date: time.Now(),
		From: &melangeProfile{
			Fingerprint: "",
			Alias:       "",
		},
		Public: true,
		Components: map[string]*melangeComponent{
			"airdispat.ch/profile/name": &melangeComponent{
				String: profileRequest.Name,
			},
			"airdispat.ch/profile/avatar": &melangeComponent{
				String: string(profileRequest.Image),
			},
			"airdispat.ch/profile/description": &melangeComponent{
				String: profileRequest.Description,
			},
		},
	}

	mail, addrs, err := msg.ToDispatch(cli.Key)
	if err != nil {
		fmt.Println("(Profile) Error converting JSON to Dispatch", err)
		return framework.Error500
	}

	if result.Profile.Value == 0 {
		// First Time Profile
		name, err := cli.PublishMessage(mail, addrs, "profile", false)
		if err != nil || name != "profile" {
			fmt.Println("(Profile) Error publishing message", err)
			return framework.Error500
		}

		_, err = c.Tables["profile"].Insert(profileRequest).Exec(c.Store)
		if err != nil {
			fmt.Println("Couldn't put profile in database.")
			return framework.Error500
		}

		result.Profile = gdb.ForeignKey(profileRequest)
		_, err = c.Tables["identity"].Update(result).Exec(c.Store)
		if err != nil {
			fmt.Println("Couldn't update id with profile", err)
			return framework.Error500
		}

	} else {
		// Update Profile
		err := cli.UpdateMessage(mail, addrs, "profile")
		if err != nil {
			fmt.Println("(Profile) Error updating message", err)
			return framework.Error500
		}

		profileRequest.Id = gdb.PrimaryKey(result.Profile.Value)
		_, err = c.Tables["profile"].Update(profileRequest).Exec(c.Store)
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
