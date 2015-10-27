package messages

import (
	"fmt"
	"sort"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"
	"getmelange.com/backend/models"
	"getmelange.com/backend/packaging"
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

func (m *Messages) Get(req *router.Request) framework.View {
	return m.view(true, true, true)
}

func (m *Messages) Post(req *router.Request) framework.View {
	request := make(map[string]bool)
	err := req.JSON(&request)
	if err != nil {
		fmt.Println("Error decoding body (MSGS)", err)
		return framework.Error500
	}

	return m.view(request["self"], request["public"], request["received"])
}

func (m *Messages) view(self, public, received bool) framework.View {
	outputMessages, err := m.retrieveMessages(self, public, received)
	if err != nil {
		fmt.Println("Error retrieving messages", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: outputMessages,
	}
}
