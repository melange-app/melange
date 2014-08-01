package controllers

import (
	"fmt"
	"melange/app/framework"
	"melange/app/models"
	"melange/router"
	"net/http"
	"sort"
	"time"

	"airdispat.ch/message"
	"airdispat.ch/server"
	"airdispat.ch/wire"

	cache "github.com/huntaub/go-cache"
	gdb "github.com/huntaub/go-db"
)

var messageCache, publicCache *cache.Cache

// {
// 		name: "adfsasdf",
// 		date: 1451435,
// 		from: 14352435,
// 		public: true,
// 		components: {},
// 		context: {},
// }

type messageList []*melangeMessage

func (m messageList) Len() int               { return len(m) }
func (m messageList) Less(i int, j int) bool { return m[i].Date.After(m[j].Date) }
func (m messageList) Swap(i int, j int)      { m[i], m[j] = m[j], m[i] }

type melangeMessage struct {
	Name       string
	Date       time.Time
	From       string
	Public     bool
	Components map[string]*melangeComponent
	Context    map[string]string
}

type melangeComponent struct {
	Name  string
	Binary []byte
	String string
}

func translateComponents(comp message.ComponentList) map[string]*melangeComponent {
	out := make(map[string]*melangeComponent)

	for key, v := range comp {
		out[key] = &melangeComponent{
			Name: v.Name,
			String: string(v.Data),
		}
	}

	return out
}

func translateMessage(public bool, msg ...*message.Mail) []*melangeMessage {
	out := make([]*melangeMessage, len(msg))

	for i, v := range msg {
		out[i] = &melangeMessage {
			Name: "",
			Date: time.Unix(v.Header().Timestamp,0),
			From: v.Header().From.String(),
			Public: public,
			Components: translateComponents(v.Components),
			Context: nil,
		}
	}

	return out
}

// Messages Controller will download messages from the server and subscribed
// sources (with caching), and send them to the client in JSON.
type Messages struct {
	Store  *models.Store
	Tables map[string]gdb.Table
}

// Handle will download the messages and transform them into JSON.
func (m *Messages) Handle(req *http.Request) framework.View {
	dap, err := CurrentDAPClient(m.Store, m.Tables["identity"])
	if err != nil {
		return err
	}

	router := &router.Router{
		Origin: dap.Key,
		TrackerList: []string{
			"localhost:2048",
		},
	}

	// request := make(map[string]interface{})
	// json, err := DecodeJSONBody(req, &request)
	// if err != nil {
	// 	fmt.Println("Error decoding body.", err)
	// 	return framework.Error500
	// }
	//
	// if _, ok := request["since"]; !ok {
	// 	return &framework.HTTPError{
	// 		ErrorCode: 400,
	// 		Message:   "Since is required.",
	// 	}
	// }

	since := uint64(0)

	outputMessages := make(messageList, 0)

	// Download Alerts
	messages, realErr := dap.DownloadMessages(since, true)
	if realErr != nil {
		fmt.Println("Error downloading messages", err)
		return framework.Error500
	}

	// Get Messages from Alerts
	for _, v := range messages {
		data, typ, h, realErr := v.Message.Reconstruct(dap.Key, false)
		if realErr != nil || typ != wire.MessageDescriptionCode {
			continue
		}

		desc, realErr := server.CreateMessageDescriptionFromBytes(data, h)
		if realErr != nil {
			continue
		}

		// TODO: h.From _MUST_ be the server key, not the client key.
		mail, realErr := downloadMessage(router, desc.Name, dap.Key, h.From.String(), desc.Location)
		if realErr != nil {
			continue
		}

		outputMessages = append(outputMessages, translateMessage(false, mail)...)
	}

	// Download Public Messages
	var s []*models.Address
	realErr = m.Tables["address"].Get().Where("subscribed", true).All(m.Store, &s)
	if realErr != nil {
		fmt.Println("Error getting contacts", realErr)
		return framework.Error500
	}

	for _, v := range s {
		var msg []*message.Mail
		list, stale := publicCache.Get(v.Fingerprint)
		if !stale {
			msg = list.([]*message.Mail)
		} else {
			var realErr error
			msg, realErr = downloadPublicMail(router, since, dap.Key, v.Fingerprint)
			if realErr != nil {
				fmt.Println("Error getting public mail", realErr)
				return framework.Error500
			}
			publicCache.Store(v.Fingerprint, msg)
		}
		outputMessages = append(outputMessages, translateMessage(true, msg...)...)
	}

	sort.Sort(outputMessages)

	return &framework.JSONView{
		Content: outputMessages,
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
	return &framework.HTTPError{
		ErrorCode: 504,
		Message:   "Not implemented.",
	}
}

// NewMessage will create and send a new message either with or without
// an alert.
type NewMessage struct {
	Store  *models.Store
	Tables map[string]gdb.Table
}

// Handle will publish the message on the server, then alert the other party.
func (m *NewMessage) Handle(req *http.Request) framework.View {
	return &framework.HTTPError{
		ErrorCode: 504,
		Message:   "Not implemented.",
	}
}
