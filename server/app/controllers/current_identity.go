package controllers

import (
	"fmt"
	"net/http"

	"getmelange.com/app/framework"
	"getmelange.com/app/models"

	gdb "github.com/huntaub/go-db"
)

// CurrentIdentity manages a REST-ful service for updating and retrieving the
// currently utilized identity.
type CurrentIdentity struct {
	Tables map[string]gdb.Table
	Store  *models.Store
}

// Handle will check if the request is POST or GET and either update the
// current identity or return it.
func (i *CurrentIdentity) Handle(req *http.Request) framework.View {
	if req.Method == "POST" {
		request := make(map[string]interface{})
		err := DecodeJSONBody(req, &request)
		if err != nil {
			fmt.Println("Cannot decode body", err)
			return framework.Error500
		}

		fingerprint, ok := request["fingerprint"]
		if !ok {
			return &framework.HTTPError{
				ErrorCode: 400,
				Message:   "Fingerprint is required.",
			}
		}

		err = i.Store.Set("current_identity", fingerprint.(string))
		if err != nil {
			fmt.Println("Error storing current_identity.", err)
			return framework.Error500
		}

		invalidateCaches()

		return &framework.JSONView{
			Content: map[string]interface{}{
				"error": false,
			},
		}

	} else if req.Method == "GET" {
		result, err := CurrentIdentityOrError(i.Store, i.Tables["identity"])
		if err != nil {
			return err
		}

		return &framework.JSONView{
			Content: result,
		}
	}
	return &framework.HTTPError{
		ErrorCode: 405,
		Message:   "Method not allowed.",
	}
}
