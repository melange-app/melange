package messages

import (
	"fmt"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"
	"getmelange.com/backend/models/messages"
)

// UpdateMessage will take a request with a messageId and a newMessage and
// change the server to point messageId to newMessage.
type UpdateMessage struct{}

// Handle will decode the JSON request and alert the server.
func (m *UpdateMessage) Post(req *router.Request) framework.View {
	msg := &messages.JSONMessage{}
	err := req.JSON(&msg)
	if err != nil {
		fmt.Println("Cannot decode message.", err)
		return framework.Error500
	}

	err = req.Environment.Manager.UpdateMessage(msg)
	if err != nil {
		fmt.Println("Error updating message", msg.Name, err)
		return framework.Error500
	}

	// Update Database

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "OK",
	}
}

var _ router.PostHandler = &UpdateMessage{}
