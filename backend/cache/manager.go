package cache

import (
	"fmt"
	"strings"
	"time"

	"airdispat.ch/identity"
	"airdispat.ch/message"

	"getmelange.com/backend/models"
	"getmelange.com/backend/models/db"
	mIdentity "getmelange.com/backend/models/identity"
	"getmelange.com/backend/models/messages"

	"getmelange.com/backend/connect"
	"getmelange.com/backend/packaging"

	gdb "github.com/huntaub/go-db"
)

const (
	logManager = "[MANAGER]"
)

// Manager is the middleman between the API module and the Connect
// module. It provides a cached version of messages when requested and
// adds messages to the cache as they come in.
type Manager struct {
	Tables *db.Tables
	Store  *models.Store

	Client *connect.Client

	Fetcher *Fetcher
	Cache   *Store

	Identity *mIdentity.Identity
	Packager *packaging.Packager

	HasIdentity bool
}

// CreateManager will build a fully initialized Manager object (with
// fully initialized Fetcher and Stores).
func CreateManager(tables *db.Tables, store *models.Store, pack *packaging.Packager) (*Manager, error) {
	cache := CreateStore()

	m := &Manager{
		Tables:   tables,
		Store:    store,
		Client:   nil,
		Fetcher:  nil,
		Cache:    cache,
		Identity: nil,
		Packager: pack,
	}

	current, err := m.currentIdentity()
	if err != nil {
		return nil, err
	}

	if current != nil {
		m.loadIdentity(current)
	}

	return m, nil
}

func (m *Manager) UpdateMessage(msg *messages.JSONMessage) error {
	mail, to, err := msg.ToDispatch(m.Client)
	if err != nil {
		return err
	}

	// Get the current version of the message in the database.
	currentMessage := &messages.Message{}
	err = m.Tables.Message.Get().Where("name", msg.Name).One(m.Store, &currentMessage)
	if err != nil {
		return err
	}

	// Update the message in the database.
	modelMessage, modelComponents := msg.ToModel(m.Client)
	modelMessage.Id = currentMessage.Id

	_, err = m.Tables.Message.Update(modelMessage).Exec(m.Store)
	if err != nil {
		return err
	}

	// Remove the old components.
	currentComponents := new([]*message.Component)
	err = m.Tables.Component.Get().Where("message_id", currentMessage.Id).
		All(m.Store, currentComponents)
	for _, v := range *currentComponents {
		_, err := m.Tables.Component.Delete(v).Exec(m.Store)
		if err != nil {
			return err
		}
	}

	// Add the new version of the components into the DB.
	for _, v := range modelComponents {
		v.Message = gdb.ForeignKey(modelMessage)
		_, err = m.Tables.Component.Insert(v).Exec(m.Store)
		if err != nil {
			fmt.Println("Couldn't save component", v.Name, err)
		}
	}

	return m.Client.UpdateMessage(mail, to, msg.Name)
}

// PublishMessage takes a JSONMessage object and publishes it through
// the connect.Client.
func (m *Manager) PublishMessage(msg *messages.JSONMessage) error {
	mail, to, err := msg.ToDispatch(m.Client)
	if err != nil {
		return err
	}

	// Add the message to the database.
	modelMessage, modelComponents := msg.ToModel(m.Client)
	_, err = m.Tables.Message.Insert(modelMessage).Exec(m.Store)
	if err != nil {
		return err
	}

	// Store each of the components in the database.
	for _, v := range modelComponents {
		v.Message = gdb.ForeignKey(modelMessage)
		_, err = m.Tables.Component.Insert(v).Exec(m.Store)
		if err != nil {
			fmt.Println("Couldn't save component", v.Name, err)
		}
	}

	// Publish Message
	// *Mail, to []*identity.Address, name string, alert bool
	name, err := m.Client.PublishMessage(mail, to, msg.Name, !msg.Public)
	if err != nil {
		return err
	}

	// Send alerts for messages that are not public.
	if !msg.Public {
		var errs []error

		for _, v := range msg.To {
			err = m.Client.SendAlert(name, v.Alias)
			if err != nil {
				logError(logManager, "Received error sending alert.", err)
				errs = append(errs, err)
			}
		}

		// Return the first error if one occurred. This may be
		// disingenious as an alert may have already been sent
		// to some parties.
		if len(errs) > 0 {
			return errs[0]
		}
	}

	go func() {
		msg.Self = true

		// Load our sending profile.
		myProfile, myAlias, _ := m.currentProfile()
		msg.From = &messages.JSONProfile{
			Name:        myProfile.Name,
			Avatar:      myProfile.Image,
			Alias:       myAlias.String(),
			Fingerprint: m.Client.Origin.Address.String(),
		}

		// Send the new message to interested parties.
		m.Fetcher.notifyObservers(msg)
	}()

	return nil
}

// GetMessage will return a message that is either in the cache or
// fetched and added to the cache.
func (m *Manager) GetMessage(alias string, name string) (*messages.JSONMessage, error) {
	msg, found := m.Cache.RetrieveMessage(alias, name)
	if !found {
		return m.Fetcher.addMessage(
			name,  // Name of the message to fetch
			alias, // Author Alias
			"",    // Server Location Alias
			nil,   // Context
		)
	}

	return msg, nil
}

// GetPublic will return a list of messages published by a user after
// the date `since`.
func (m *Manager) GetPublic(since uint64, addr *mIdentity.Address) ([]*messages.JSONMessage, error) {
	msgs, found := m.Cache.RetrievePublic(addr.Alias)
	if !found {
		return m.Fetcher.addPublic(since, addr)
	}

	// Filter out messages that may have occurred before the
	// requested date.
	sinceT := time.Unix(int64(since), 0)
	var output []*messages.JSONMessage
	for _, m := range msgs {
		if sinceT.Before(m.Date) {
			output = append(output, m)
		}
	}
	return output, nil
}

// GetSentMessages will return a list of messages that were sent by the user after the date `since`.
func (m *Manager) GetSentMessages(since uint64) ([]*messages.JSONMessage, error) {
	msgs := make([]*messages.Message, 0)
	err := m.Tables.Message.Get().Where("from", m.Client.Origin.Address.String()).All(m.Store, &msgs)
	if err != nil {
		return nil, err
	}

	myProfile, myAlias, err := m.currentProfile()
	if err != nil {
		return nil, err
	}

	myInformation := &messages.TranslationMe{
		Profile: myProfile,
		Alias:   myAlias,
	}

	var outputMessages []*messages.JSONMessage
	for _, modelMessage := range msgs {
		// Build a list of profiles that we need
		var toAddress []*identity.Address
		for _, addr := range strings.Split(modelMessage.To, ",") {
			toAddress = append(toAddress, &identity.Address{
				Alias: addr,
			})
		}

		serializedMessage, err := m.Fetcher.Translator.Request(&messages.TranslationRequest{
			Model:    modelMessage,
			Me:       myInformation,
			Profiles: m.Fetcher.buildProfiles(toAddress...),
		})

		if err != nil {
			logError(logManager, "Received error translating model.", err)
		}
		outputMessages = append(outputMessages, serializedMessage)
	}

	return outputMessages, nil
}

// GetAllMessages will send all of the messages currently in the cache
// to the requesting channel so long as that channel is already an
// observer of the majorFetcher.
func (m *Manager) GetAllMessages(msgChan chan *messages.JSONMessage) {
	var newData messages.JSONMessageList

	if !m.Fetcher.Initial {
		// If we haven't done the initial fetch, go ahead and do that.
		m.Fetcher.CheckForMessages(0)
	} else {
		// Otherwise, return all the messages we currently
		// have cached.
		newData = messages.JSONMessageList(
			m.Cache.RetrieveAll())
	}

	// Get all of the sent messages thusfar.
	data, err := m.GetSentMessages(0)
	if err != nil {
		logError(logManager, "Received error getting sent messages", err)
	}
	newData = append(newData, data...)


	// Give the data to the channel. Note: This will block unless
	// the channel is continuously reading messages.
	for _, v := range newData {
		msgChan <- v
	}
}

// GetProfile will attempt to fetch a profile out of the cache. If it
// isn't found, it will fetch it (or return the default).
func (m *Manager) GetProfile(of string) (*messages.JSONProfile, error) {
	return m.Fetcher.getProfile(of)
}

// currentProfile constructs the current user profile and alias
// objects.
func (m *Manager) currentProfile() (*mIdentity.Profile, *mIdentity.Alias, error) {
	myAlias := &mIdentity.Alias{}
	err := m.Tables.Alias.Get().Where("identity", m.Identity.Id).One(m.Store, myAlias)
	if err != nil {
		return nil, nil, err
	}

	myProfile := &mIdentity.Profile{}
	err = m.Identity.Profile.One(m.Store, myProfile)
	if err != nil || myProfile.Id == 0 {
		fmt.Println(logManager, "Received error getting my profile", err)
		myProfile = &mIdentity.Profile{
			Name:  myAlias.String(),
			Image: m.Fetcher.Translator.DefaultProfileImage(m.Client.Origin.Address.Alias),
		}
	}

	if strings.Contains(myProfile.Image, "@") {
		myProfile.Image = fmt.Sprintf("http://data.local.getmelange.com:7776/%s", myProfile.Image)
	}

	return myProfile, myAlias, nil
}
