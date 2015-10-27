package messages

import (
	"fmt"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"
	"getmelange.com/backend/models"
)

// UpdateMessage will take a request with a messageId and a newMessage and
// change the server to point messageId to newMessage.
type UpdateMessage struct{}

// Handle will decode the JSON request and alert the server.
func (m *UpdateMessage) Post(req *router.Request) framework.View {
	msg := &models.JSONMessage{}
	err := req.JSON(&msg)
	if err != nil {
		fmt.Println("Cannot decode message.", err)
		return framework.Error500
	}

	// Get current DAP Client
	dap, fErr := CurrentDAPClient(req.Environment.Settings, req.Environment.Tables()["identity"])
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
	err = m.Tables["message"].Get().Where("name", msg.Name).One(req.Environment.Settings, currentMsg)
	if err != nil {
		fmt.Println("That message doesn't exist.", err)
		return framework.Error500
	}

	modelMsg.Id = currentMsg.Id

	_, err = m.Tables["message"].Update(modelMsg).Exec(req.Environment.Settings)
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

var _ router.PostHandler = &UpdateMessage{}
