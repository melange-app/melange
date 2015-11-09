package messages

import (
	"fmt"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"
	"getmelange.com/backend/models/messages"
	"getmelange.com/backend/packaging"
)

// NewMessage will create and send a new message either with or without
// an alert.
type NewMessage struct {
	Packager *packaging.Packager
}

// Handle will publish the message on the server, then alert the other party.
func (m *NewMessage) Post(req *router.Request) framework.View {
	msg := &messages.JSONMessage{}
	err := req.JSON(&msg)
	if err != nil {
		fmt.Println("Cannot decode message.", err)
		return framework.Error500
	}

	err = req.Environment.Manager.PublishMessage(msg)
	if err != nil {
		fmt.Println("Error publishing message", err)
		return framework.Error500
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "Done!",
	}
}

var _ router.PostHandler = &NewMessage{}
