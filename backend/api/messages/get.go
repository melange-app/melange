package messages

import (
	"fmt"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/models"
	"getmelange.com/backend/models/messages"
	"getmelange.com/router"
)

type getMessageRequest struct {
	Name       string `json:"name"`
	Alias      string `json:"alias"`
	OnlyPublic bool   `json:"onlyPublic"`
	Since      uint64 `json:"since"`
}

// GetMessage controller will download a single message as specified by
// Alias and Name
type GetAllMessagesAt struct {
	Store  *models.Store
	Tables map[string]gdb.Table
}

// Handle will download the messages and transform them into JSON.
func (m *GetAllMessagesAt) Post(req *router.Request) framework.View {
	request := &getMessageRequest{}
	realErr := DecodeJSONBody(req, &request)
	if realErr != nil {
		fmt.Println("Error decoding body (MSGS).", realErr)
		return framework.Error500
	}

	id, frameErr := CurrentIdentityOrError(m.Store, m.Tables["identity"])
	if frameErr != nil {
		return frameErr
	}

	client, err := DAPClientFromID(id, m.Store)
	if err != nil {
		fmt.Println("Couldn't construct DAPClient", err)
		return framework.Error500
	}

	router := &router.Router{
		Origin: client.Key,
		TrackerList: []string{
			"localhost:2048",
		},
	}

	manager := &messages.MessageManager{
		Router:   router,
		Client:   client,
		Tables:   m.Tables,
		Store:    m.Store,
		Identity: id,
	}

	msgs, err := manager.GetPublic(request.Since, &models.Address{
		Alias: request.Alias,
	})

	if err != nil {
		fmt.Println("Couldn't get public main", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: msgs,
	}
}

// GetMessage controller will download a single message as specified by
// Alias and Name
type GetMessage struct {
	Store  *models.Store
	Tables map[string]gdb.Table
}

// Handle will download the messages and transform them into JSON.
func (m *GetMessage) Post(req *router.Request) framework.View {
	request := &getMessageRequest{}
	realErr := DecodeJSONBody(req, &request)
	if realErr != nil {
		fmt.Println("Error decoding body (MSGS).", realErr)
		return framework.Error500
	}

	id, frameErr := CurrentIdentityOrError(m.Store, m.Tables["identity"])
	if frameErr != nil {
		return frameErr
	}

	client, err := DAPClientFromID(id, m.Store)
	if err != nil {
		fmt.Println("Couldn't construct DAPClient", err)
		return framework.Error500
	}

	router := &router.Router{
		Origin: client.Key,
		TrackerList: []string{
			"localhost:2048",
		},
	}

	manager := &messages.MessageManager{
		Router:   router,
		Client:   client,
		Tables:   m.Tables,
		Store:    m.Store,
		Identity: id,
	}

	msg, err := manager.GetMessage(request.Alias, request.Name)
	if err != nil {
		fmt.Println("Error getting message", request.Name, err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: msg,
	}
}
