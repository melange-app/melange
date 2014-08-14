package controllers

import (
	"errors"
	"fmt"
	"melange/app/framework"
	"melange/app/models"
	"melange/router"
	"net/http"
	"sort"
	"time"

	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"
	"airdispat.ch/server"
	"airdispat.ch/wire"

	cache "github.com/huntaub/go-cache"
	gdb "github.com/huntaub/go-db"
)

var (
	cacheDuration             = 2 * time.Minute
	messageCache, publicCache = cache.NewCache(cacheDuration), cache.NewCache(cacheDuration)
	profileCache              = cache.NewCache(cacheDuration)
)

func invalidateCaches() {
	messageCache, publicCache = cache.NewCache(cacheDuration), cache.NewCache(cacheDuration)
	profileCache = cache.NewCache(cacheDuration)
}

type messageList []*melangeMessage

func (m messageList) Len() int               { return len(m) }
func (m messageList) Less(i int, j int) bool { return m[i].Date.After(m[j].Date) }
func (m messageList) Swap(i int, j int)      { m[i], m[j] = m[j], m[i] }

type melangeMessage struct {
	Name       string                       `json:"name"`
	Date       time.Time                    `json:"date"`
	From       melangeProfile               `json:"from"`
	To         []string                     `json:"to"`
	Public     bool                         `json:"public"`
	Components map[string]*melangeComponent `json:"components"`
	Context    map[string]string            `json:"context"`
}

type melangeComponent struct {
	Binary []byte `json:"binary"`
	String string `json:"string"`
}

type melangeProfile struct {
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Alias       string `json:"alias"`
	Fingerprint string `json:"fingerprint"`
}

func (m melangeProfile) UnmarshalJSON(data []byte) error {
	// Data should be a string straight up.
	fingerprint := string(data)
	m.Fingerprint = fingerprint
	return nil
}

func (m *melangeMessage) ToDispatch(from *identity.Identity) (*message.Mail, []*identity.Address, error) {
	r := router.Router{
		Origin: from,
	}

	addrs := make([]*identity.Address, len(m.To))
	for i, v := range m.To {
		var err error
		addrs[i], err = r.LookupAlias(v, routing.LookupTypeMAIL)
		if err != nil {
			return nil, nil, err
		}
	}

	mail := message.CreateMail(from.Address, m.Date, addrs...)

	for key, v := range m.Components {
		mail.Components.AddComponent(message.Component{
			Name: key,
			Data: []byte(v.String),
		})
	}

	return mail, addrs, nil
}

func translateComponents(comp message.ComponentList) map[string]*melangeComponent {
	out := make(map[string]*melangeComponent)

	for _, v := range comp {
		out[v.Name] = &melangeComponent{
			String: string(v.Data),
		}
	}

	return out
}

func translateMessageWithContext(r routing.Router, from *identity.Identity, public bool, context map[string][]byte, msg *message.Mail) []*melangeMessage {
	obj := translateMessage(r, from, public, msg)

	out := make(map[string]string)
	for key, v := range context {
		out[key] = string(v)
	}

	obj[0].Context = out

	return obj
}

func translateMessage(r routing.Router, from *identity.Identity, public bool, msg ...*message.Mail) []*melangeMessage {
	out := make([]*melangeMessage, len(msg))

	for i, v := range msg {
		var profile melangeProfile
		obj, stale := profileCache.Get(v.Header().Alias)
		if !stale {
			profile = obj.(melangeProfile)
		} else {
			var err error

			profile, err = translateProfile(r, from, v.Header().From.String(), v.Header().Alias)
			if err != nil {
				fmt.Println("Couldn't get profile", err)

				name := v.Header().From.String()
				if v.Header().Alias != "" {
					name = v.Header().Alias
				}

				profile = melangeProfile{
					Name:        name,
					Fingerprint: v.Header().From.String(),
					Avatar:      "http://placehold.it/404", // haha
					Alias:       v.Header().Alias,
				}
			} else {
				profileCache.Store(v.Header().Alias, profile)
			}
		}

		out[i] = &melangeMessage{
			Name:       "",
			Date:       time.Unix(v.Header().Timestamp, 0),
			From:       profile,
			Public:     public,
			Components: translateComponents(v.Components),
			Context:    nil,
		}
	}

	return out
}

func translateProfile(r routing.Router, from *identity.Identity, fp string, alias string) (melangeProfile, error) {
	if alias == "" {
		return melangeProfile{}, errors.New("Can't get profile without alias support.")
	}

	profile, err := getProfile(r, from, fp, alias)
	if err != nil {
		return melangeProfile{}, err
	}

	name := profile.Components.GetStringComponent("airdispat.ch/profile/name")
	if name == "" {
		name = alias
	}

	avatar := profile.Components.GetStringComponent("airdispat.ch/profile/avatar")
	if avatar == "" {
		avatar = "http://placehold.it/404"
	}

	return melangeProfile{
		Name:        name,
		Avatar:      avatar,
		Alias:       alias,
		Fingerprint: fp,
	}, nil
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
		fmt.Println("Error downloading messages", realErr)
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
			fmt.Println("Can't create message description", realErr)
			continue
		}

		// TODO: h.From _MUST_ be the server key, not the client key.
		mail, realErr := downloadMessage(router, desc.Name, dap.Key, h.From.String(), desc.Location)
		if realErr != nil {
			fmt.Println("Got error downloading message", desc.Name, realErr)
			continue
		}

		outputMessages = append(outputMessages, translateMessageWithContext(router, dap.Key, false, v.Context, mail)...)
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
			msg, realErr = downloadPublicMail(router, since, dap.Key, v)
			if realErr != nil {
				fmt.Println("Error getting public mail", realErr)
				return framework.Error500
			}
			publicCache.Store(v.Fingerprint, msg)
		}
		outputMessages = append(outputMessages, translateMessage(router, dap.Key, true, msg...)...)
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
	msg := &melangeMessage{}
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

	err = dap.UpdateMessage(mail, to, msg.Name)
	if err != nil {
		fmt.Println("Can't update message.", err)
		return framework.Error500
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "OK",
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
	msg := &melangeMessage{}
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

	// Publish Message
	// *Mail, to []*identity.Address, name string, alert bool
	name, err := dap.PublishMessage(mail, to, msg.Name, !msg.Public)

	if !msg.Public {
		fmt.Println("Sending alerts...")
		// Send Alert
		var errs []error

		r := &router.Router{
			Origin: dap.Key,
		}

		for _, v := range msg.To {
			fmt.Println("Sending alert to", v)
			err = sendAlert(r, name, dap.Key, v, dap.Server.Alias)
			if err != nil {
				fmt.Println("Got error sending alert", err)
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			return framework.Error500
		}
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "Done!",
	}
}
