package controllers

import (
	"fmt"
	"net/http"
	"sort"

	"getmelange.com/app/framework"
	"getmelange.com/app/messages"
	"getmelange.com/app/models"
	"getmelange.com/app/packaging"
	"getmelange.com/router"
	gdb "github.com/huntaub/go-db"
)

// Messages Controller will download messages from the server and subscribed
// sources (with caching), and send them to the client in JSON.
type Messages struct {
	Packager *packaging.Packager
	Store    *models.Store
	Tables   map[string]gdb.Table
}

func (m *Messages) retrieveMessages(self, public, received bool) ([]*models.JSONMessage, error) {
	manager, err := constructManager(m.Store, m.Tables, m.Packager)
	if err != nil {
		return nil, err
	}

	since := uint64(0)

	var outputMessages models.JSONMessageList

	if received {
		outputMessages = append(outputMessages, manager.GetPrivateMessages(since, nil)...)
	}

	if self {
		msgs, err := manager.GetSentMessages(since, nil)
		if err != nil {
			return nil, err
		}

		outputMessages = append(outputMessages, msgs...)
	}

	// Download Public Messages
	if public {
		outputMessages = append(outputMessages, manager.GetPublicMessages(since, nil)...)
	}

	sort.Sort(outputMessages)

	fmt.Println("Finished Retrieving Messages")

	return outputMessages, nil
}

// Handle will download the messages and transform them into JSON.
func (m *Messages) Handle(req *http.Request) framework.View {
	request := make(map[string]bool)
	if req.Method == "GET" {
		request["self"] = true
		request["public"] = true
		request["received"] = true
	} else {
		realErr := DecodeJSONBody(req, &request)
		if realErr != nil {
			fmt.Println("Error decoding body (MSGS).", realErr)
			return framework.Error500
		}
	}

	outputMessages, err := m.retrieveMessages(request["self"], request["public"], request["received"])
	if err != nil {
		fmt.Println("Error retrieving messages", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: outputMessages,
	}
}

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
func (m *GetAllMessagesAt) Handle(req *http.Request) framework.View {
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
func (m *GetMessage) Handle(req *http.Request) framework.View {
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

// UpdateMessage will take a request with a messageId and a newMessage and
// change the server to point messageId to newMessage.
type UpdateMessage struct {
	Store  *models.Store
	Tables map[string]gdb.Table
}

// Handle will decode the JSON request and alert the server.
func (m *UpdateMessage) Handle(req *http.Request) framework.View {
	msg := &models.JSONMessage{}
	err := DecodeJSONBody(req, &msg)
	if err != nil {
		fmt.Println("Cannot decode message.", err)
		return framework.Error500
	}

	// Get current DAP Client
	dap, fErr := CurrentDAPClient(m.Store, m.Tables["identity"])
	if fErr != nil {
		return fErr
	}

	mail, to, err := msg.ToDispatch(dap.Key)
	if err != nil {
		fmt.Println("Couldn't convert JSON to Dispatcher", err)
		return framework.Error500
	}

	modelMsg, modelComp := msg.ToModel(dap.Key)

	currentMsg := &models.Message{}
	err = m.Tables["message"].Get().Where("name", msg.Name).One(m.Store, currentMsg)
	if err != nil {
		fmt.Println("That message doesn't exist.", err)
		return framework.Error500
	}

	modelMsg.Id = currentMsg.Id

	_, err = m.Tables["message"].Update(modelMsg).Exec(m.Store)
	if err != nil {
		fmt.Println("Couldn't update the message.", err)
		return framework.Error500
	}

	currentComp := make([]*models.Component, 0)
	err = m.Tables["component"].Get().Where("message", modelMsg.Id).All(m.Store, &currentComp)
	if err != nil {
		fmt.Println("Couldn't get the components.", err)
		return framework.Error500
	}

	for _, v := range currentComp {
		_, err = m.Tables["component"].Delete(currentComp).Exec(m.Store)
		if err != nil {
			fmt.Println("Couldn't delete the component", v.Name, err)
		}
	}

	for _, v := range modelComp {
		v.Message = gdb.ForeignKey(modelMsg)
		_, err = m.Tables["component"].Insert(v).Exec(m.Store)
		if err != nil {
			fmt.Println("Couldn't save component", v.Name, err)
		}
	}

	err = dap.UpdateMessage(mail, to, msg.Name)
	if err != nil {
		fmt.Println("Can't update message.", err)
		return framework.Error500
	}

	// Update Database

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "OK",
	}
}

// NewMessage will create and send a new message either with or without
// an alert.
type NewMessage struct {
	Store    *models.Store
	Tables   map[string]gdb.Table
	Packager *packaging.Packager
}

// Handle will publish the message on the server, then alert the other party.
func (m *NewMessage) Handle(req *http.Request) framework.View {
	msg := &models.JSONMessage{}
	err := DecodeJSONBody(req, &msg)
	if err != nil {
		fmt.Println("Cannot decode message.", err)
		return framework.Error500
	}

	manager, err := constructManager(m.Store, m.Tables, m.Packager)
	if err != nil {
		fmt.Println("Error creating manager", err)
		return framework.Error500
	}

	err = manager.PublishMessage(msg)
	if err != nil {
		fmt.Println("Error publishing message", err)
		return framework.Error500
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "Done!",
	}
}
