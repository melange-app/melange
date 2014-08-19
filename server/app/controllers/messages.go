package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"getmelange.com/app/framework"
	"getmelange.com/app/models"
	"getmelange.com/router"

	adErrors "airdispat.ch/errors"
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
	From       *melangeProfile              `json:"from"`
	To         []*melangeProfile            `json:"to"`
	Public     bool                         `json:"public"`
	Self       bool                         `json:"self"`
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

func (m *melangeMessage) ToModel(from *identity.Identity) (*models.Message, []*models.Component) {
	toAddrs := ""
	for i, v := range m.To {
		if i != 0 {
			toAddrs += ","
		}
		toAddrs += v.Alias
	}

	message := &models.Message{
		Name: m.Name,
		// Address
		To:   toAddrs,
		From: from.Address.String(),
		// Meta
		Date:     m.Date.Unix(),
		Incoming: false,
		Alert:    !m.Public,
	}

	components := make([]*models.Component, len(m.Components))
	i := 0
	for key, v := range m.Components {
		components[i] = &models.Component{
			Name: key,
		}

		if len(v.Binary) == 0 {
			components[i].Data = []byte(v.String)
		} else {
			components[i].Data = v.Binary
		}

		i++
	}

	return message, components
}

func (m *melangeMessage) ToDispatch(from *identity.Identity) (*message.Mail, []*identity.Address, error) {
	r := router.Router{
		Origin: from,
	}

	addrs := make([]*identity.Address, len(m.To))
	for i, v := range m.To {
		var err error
		addrs[i], err = r.LookupAlias(v.Alias, routing.LookupTypeMAIL)
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
		var profile *melangeProfile
		obj, stale := profileCache.Get(v.Header().Alias)
		if !stale {
			profile = obj.(*melangeProfile)
		} else {
			var err error

			profile, err = translateProfile(r, from, v.Header().From.String(), v.Header().Alias)
			if err != nil {
				fmt.Println("Couldn't get profile", err)

				name := v.Header().From.String()
				if v.Header().Alias != "" {
					name = v.Header().Alias
				}

				profile = &melangeProfile{
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

func translateProfile(r routing.Router, from *identity.Identity, fp string, alias string) (*melangeProfile, error) {
	if alias == "" {
		return nil, errors.New("Can't get profile without alias support.")
	}

	noImage := "http://placehold.it/404"

	profile, err := getProfile(r, from, fp, alias)
	if err != nil {
		switch t := err.(type) {
		case (*adErrors.Error):
			if t.Code == 5 {
				return &melangeProfile{
					Name:        alias,
					Avatar:      noImage,
					Alias:       alias,
					Fingerprint: fp,
				}, nil
			}
			return nil, err
		case error:
			return nil, err
		}
	}

	name := profile.Components.GetStringComponent("airdispat.ch/profile/name")
	if name == "" {
		name = alias
	}

	avatar := profile.Components.GetStringComponent("airdispat.ch/profile/avatar")
	if avatar == "" {
		avatar = noImage
	}

	return &melangeProfile{
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

	id, err := CurrentIdentityOrError(m.Store, m.Tables["identity"])
	if err != nil {
		return err
	}

	dap, realErr := DAPClientFromID(id, m.Store)
	if realErr != nil {
		fmt.Println("Couldn't construct DAPClient", err)
		return framework.Error500
	}

	router := &router.Router{
		Origin: dap.Key,
		TrackerList: []string{
			"localhost:2048",
		},
	}

	since := uint64(0)

	outputMessages := make(messageList, 0)

	// Download Alerts
	messages, realErr := dap.DownloadMessages(since, true)
	if realErr != nil {
		fmt.Println("Error downloading messages", realErr)
		return framework.Error500
	}

	if request["received"] {
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
	}

	if request["self"] {
		msgs := make([]*models.Message, 0)
		realErr = m.Tables["message"].Get().Where("from", dap.Key.Address.String()).All(m.Store, &msgs)
		if realErr != nil {
			fmt.Println("Unable to get self messages.", realErr)
			return framework.Error500
		}

		myAlias := &models.Alias{}
		realErr = m.Tables["alias"].Get().Where("identity", id.Id).One(m.Store, myAlias)
		if realErr != nil {
			fmt.Println("Unable to get my profile.", realErr)
			return framework.Error500
		}

		myProfile := &models.Profile{}
		realErr = id.Profile.One(m.Store, myProfile)
		if realErr != nil || myProfile.Id == 0 {
			fmt.Println("Unable to get my profile.", realErr)
			myProfile = &models.Profile{
				Name:  myAlias.String(),
				Image: "http://placehold.it/400",
			}
		}

		for _, v := range msgs {
			comps := make([]*models.Component, 0)
			realErr = m.Tables["component"].Get().Where("message", v.Id).All(m.Store, &comps)
			if realErr != nil {
				fmt.Println("Unable to get self message components", v.Name, err)
			}

			mlgComps := make(map[string]*melangeComponent)
			for _, c := range comps {
				mlgComps[c.Name] = &melangeComponent{
					Binary: c.Data,
					String: string(c.Data),
				}
			}

			// Download Profile
			toAddrs := strings.Split(v.To, ",")

			profiles := make([]*melangeProfile, 0)
			for _, j := range toAddrs {
				if j == "" {
					continue
				}

				_, addr, err := getAddresses(router, &models.Address{
					Alias: j,
				})
				if err != nil {
					fmt.Println("Couldn't get fp for", j, err)
					continue
				}

				p, err := translateProfile(router, dap.Key, addr.String(), j)
				if err != nil {
					fmt.Println("Couldn't get profile for", j, err)
				}

				profiles = append(profiles, p)
			}

			outputMessages = append(outputMessages, &melangeMessage{
				Name: v.Name,
				Date: time.Unix(v.Date, 0),
				// To and From Info
				From: &melangeProfile{
					Name:        myProfile.Name,
					Avatar:      myProfile.Image,
					Alias:       myAlias.String(),
					Fingerprint: dap.Key.Address.String(),
				},
				To: profiles,
				// Components
				Components: mlgComps,
				// Meta
				Self:   true,
				Public: !v.Alert,
			})
		}
	}

	// Download Public Messages
	if request["public"] {
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
					switch t := realErr.(type) {
					case *adErrors.Error:
						if t.Code == 5 {
							// No public mail available for that user.
							// No alert needed.
							continue
						} else {
							fmt.Println("Error getting public mail", realErr)
							return framework.Error500
						}
					case error:
						fmt.Println("Error getting public mail", realErr)
						return framework.Error500
					}
				}
				publicCache.Store(v.Fingerprint, msg)
			}
			outputMessages = append(outputMessages, translateMessage(router, dap.Key, true, msg...)...)
		}
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

	// Add to Database
	modelMsg, modelComp := msg.ToModel(dap.Key)
	_, err = m.Tables["message"].Insert(modelMsg).Exec(m.Store)
	if err != nil {
		fmt.Println("Can't save message.", err)
		return framework.Error500
	}

	for _, v := range modelComp {
		v.Message = gdb.ForeignKey(modelMsg)
		_, err = m.Tables["component"].Insert(v).Exec(m.Store)
		if err != nil {
			fmt.Println("Couldn't save component", v.Name, err)
		}
	}

	// Publish Message
	// *Mail, to []*identity.Address, name string, alert bool
	name, err := dap.PublishMessage(mail, to, msg.Name, !msg.Public)
	if err != nil {
		fmt.Println("Can't publish messages")
		return framework.Error500
	}

	if !msg.Public {
		fmt.Println("Sending alerts...")
		// Send Alert
		var errs []error

		r := &router.Router{
			Origin: dap.Key,
		}

		for _, v := range msg.To {
			fmt.Println("Sending alert to", v.Alias)
			err = sendAlert(r, name, dap.Key, v.Alias, dap.Server.Alias)
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
