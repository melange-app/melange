package messages

import (
	"fmt"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"
	"getmelange.com/backend/models/identity"
)

type getMessageRequest struct {
	Name       string `json:"name"`
	Alias      string `json:"alias"`
	OnlyPublic bool   `json:"onlyPublic"`
	Since      uint64 `json:"since"`
}

// GetMessage controller will download a single message as specified by
// Alias and Name
type GetAllMessagesAt struct{}

// Post will download the messages and transform them into JSON.
func (m *GetAllMessagesAt) Post(req *router.Request) framework.View {
	request := &getMessageRequest{}
	if err := req.JSON(&request); err != nil {
		fmt.Println("Error decoding body (MSGS).", err)
		return framework.Error500
	}

	msgs, err := req.Environment.Manager.GetPublic(request.Since, &identity.Address{
		Alias: request.Alias,
	})

	if err != nil {
		fmt.Println("[MSG] Couldn't messages for", request.Alias)
		fmt.Println("[MSG]", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: msgs,
	}
}

// GetMessage controller will download a single message as specified by
// Alias and Name
type GetMessage struct{}

// Handle will download the messages and transform them into JSON.
func (m *GetMessage) Post(req *router.Request) framework.View {
	request := &getMessageRequest{}
	if err := req.JSON(&request); err != nil {
		fmt.Println("Error decoding body (MSGS).", err)
		return framework.Error500
	}

	msg, err := req.Environment.Manager.GetMessage(request.Alias, request.Name)
	if err != nil {
		fmt.Println("[MSG] Couldn't get message", request.Name, "from", request.Alias)
		fmt.Println("[MSG]", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: msg,
	}
}
