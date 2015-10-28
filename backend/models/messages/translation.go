package messages

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"airdispat.ch/message"
	"getmelange.com/backend/models/db"
	mIdentity "getmelange.com/backend/models/identity"

	gdb "github.com/huntaub/go-db"
)

// TranslationRequest represents a request to translate a message.
type TranslationRequest struct {
	Model *Message
	Me    *TranslationMe

	Message *message.Mail
	Public  bool
	Name    string

	Profiles map[string]*JSONProfile

	Context map[string][]byte
}

// TranslationMe keeps track of the object that are loaded from the
// database about the current user.
type TranslationMe struct {
	Alias   *mIdentity.Alias
	Profile *mIdentity.Profile
}

// Translator is an object that performs translation of messages to
// and from JSON.
type Translator struct {
	Tables *db.Tables
	Store  gdb.Executor
}

// CreateTranslator will make a new translator for the object.
func CreateTranslator(tables *db.Tables, store gdb.Executor) *Translator {
	return &Translator{
		Tables: tables,
		Store:  store,
	}
}

func (t *Translator) Request(req *TranslationRequest) (*JSONMessage, error) {
	if req.Message != nil {
		return t.fromMessage(req), nil
	}

	if req.Model != nil {
		return t.fromModel(req)
	}

	return nil, errors.New("translation: Bad request")
}

// fromComponents will take a list of AirDispatch components and
// turn them into the corresponding JSON values.
func (t *Translator) fromComponents(comp message.ComponentList) map[string]*JSONComponent {
	out := make(map[string]*JSONComponent)

	for _, v := range comp {
		out[v.Name] = &JSONComponent{
			String: string(v.Data),
		}
	}

	return out
}

// fromModel will change a Melange Message model into a JSON message.
func (t *Translator) fromModel(req *TranslationRequest) (*JSONMessage, error) {
	cmpTable := t.Tables.Component

	// Get all of the components associated with that message.
	comps := make([]*Component, 0)
	err := cmpTable.Get().Where("message", req.Model.Id).All(t.Store, &comps)
	if err != nil {
		return nil, err
	}

	// Translate the components into JSON.
	parsedComponents := make(map[string]*JSONComponent)
	for _, c := range comps {
		parsedComponents[c.Name] = &JSONComponent{
			Binary: c.Data,
			String: string(c.Data),
		}
	}

	toAddrs := strings.Split(req.Model.To, ",")

	// Download the profiles of all recipients of the message.
	profiles := make([]*JSONProfile, len(toAddrs))
	for i, toAddr := range toAddrs {
		profiles[i] = req.Profiles[toAddr]
	}

	myProfile := &JSONProfile{
		Name:   req.Me.Profile.Name,
		Avatar: req.Me.Profile.Image,
		Alias:  req.Me.Alias.String(),
	}

	// Export the finished product.
	return &JSONMessage{
		Name: req.Model.Name,
		Date: time.Unix(req.Model.Date, 0),

		// To and From Info
		From: myProfile,
		To:   profiles,

		// Components
		Components: parsedComponents,

		// Meta
		Self:   true,
		Public: !req.Model.Alert,
	}, nil
}

// fromMessage will convert a series of AirDispatch messages into a
// series of JSONMessages.
func (t *Translator) fromMessage(req *TranslationRequest) *JSONMessage {
	header := req.Message.Header()

	// Load the profile of the sender into the object.
	fromProfile := req.Profiles[header.From.String()]

	// Load the profile of the recipients into the object.
	toProfiles := make([]*JSONProfile, len(header.To))
	for i, addr := range header.To {
		toProfiles[i] = req.Profiles[addr.String()]
	}

	// For legacy reasons, we will set a 'random' string
	// as the name of the message if it doesn't have one.
	// TODO: Change to GUID.
	named := req.Message.Name
	if named == "" && req.Name != "" {
		named = req.Name
	} else if named == "" {
		named = fmt.Sprintf("__%d", time.Now().Unix())
	}

	// Convert the request context (binary) to strings.
	context := make(map[string]string)
	for key, v := range req.Context {
		context[key] = string(v)
	}

	// Output the object as a JSONMessage
	return &JSONMessage{
		Name:       named,
		Date:       time.Unix(req.Message.Header().Timestamp, 0),
		To:         toProfiles,
		From:       fromProfile,
		Public:     req.Public,
		Components: t.fromComponents(req.Message.Components),
		Context:    context,
	}
}

// fromProfile will return the parsed profile out of an AirDispatch
// message.
func (t *Translator) FromProfile(alias string, profile *message.Mail) *JSONProfile {
	// Provide default profile if none is downloaded.
	if profile == nil {
		return &JSONProfile{
			Name:   alias,
			Avatar: t.DefaultProfileImage(alias),
			Alias:  alias,
		}
	}

	// Extract name component.
	name := profile.Components.GetStringComponent("airdispat.ch/profile/name")
	if name == "" {
		name = alias
	}

	// Extract avatar component.
	avatar := profile.Components.GetStringComponent("airdispat.ch/profile/avatar")
	if avatar == "" {
		avatar = t.DefaultProfileImage(alias)
	}

	// Check for AirDispatch address in the avatar and convert if
	// necessary.
	if strings.Contains(avatar, "@") {
		// TODO: Load this from the environment rather than
		// hardcoded.
		dataURL := "http://data.local.getmelange.com:7776"
		avatar = fmt.Sprintf("%s/%s", dataURL, avatar)
	}

	// Return the completed profile object.
	return &JSONProfile{
		Name:   name,
		Avatar: avatar,
		Alias:  alias,
	}
}

// DefaultProfileImage provides a profile image for messages that do
// not have a hosted profile from robohash.
func (t *Translator) DefaultProfileImage(key string) string {
	return fmt.Sprintf("http://robohash.org/%s.png?bgset=bg2", key)
}

// func (t *Translator) addProfile(c *models.Contact) error {
// 	currentAddress := c.Identities[0]

// 	fp := currentAddress.Fingerprint
// 	if fp == "" {
// 		temp, err := r.LookupAlias(currentAddress.Alias, routing.LookupTypeMAIL)
// 		if err != nil {
// 			return err
// 		}
// 		fp = temp.String()

// 	}

// 	profile, err := f.fromProfile(fp, currentAddress.Alias)
// 	if err != nil {
// 		return err
// 	}

// 	c.Profile = profile
// 	return nil
// }
