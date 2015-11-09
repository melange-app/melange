package messages

import (
	"fmt"
	"sort"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"
	"getmelange.com/backend/models/messages"
)

type messageRequest struct {
	Self     bool `json:"self"`
	Public   bool `json:"public"`
	Received bool `json:"received"`
}

// Messages Controller will download messages from the server and subscribed
// sources (with caching), and send them to the client in JSON.
type Messages struct{}

func (m *Messages) retrieveMessages(mReq *messageRequest, req *router.Request) ([]*messages.JSONMessage, error) {
	since := uint64(0)

	var outputMessages messages.JSONMessageList

	manager := req.Environment.Manager

	if mReq.Received {
		// TODO: Determine if this method is actually used and
		// finish it.

		// outputMessages = append(outputMessages, manager.GetPrivateMessages(since, nil)...)
	}

	if mReq.Self {
		msgs, err := manager.GetSentMessages(since)
		if err != nil {
			return nil, err
		}

		outputMessages = append(outputMessages, msgs...)
	}

	// Download Public Messages
	if mReq.Public {
		msgs, err := manager.GetPublic(since, nil)
		if err != nil {
			return nil, err
		}

		outputMessages = append(outputMessages, msgs...)
	}

	sort.Sort(outputMessages)

	fmt.Println("Finished Retrieving Messages")

	return outputMessages, nil
}

func (m *Messages) Get(req *router.Request) framework.View {
	outputMessages, err := m.retrieveMessages(&messageRequest{
		Self:     true,
		Public:   true,
		Received: true,
	}, req)
	if err != nil {
		fmt.Println("Error retrieving messages", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: outputMessages,
	}
}

func (m *Messages) Post(req *router.Request) framework.View {
	request := &messageRequest{}
	err := req.JSON(&request)
	if err != nil {
		fmt.Println("Error decoding body (MSGS)", err)
		return framework.Error500
	}

	outputMessages, err := m.retrieveMessages(request, req)
	if err != nil {
		fmt.Println("Error retrieving messages", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: outputMessages,
	}
}
