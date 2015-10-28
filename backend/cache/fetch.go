package cache

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"getmelange.com/backend/models/db"
	mIdentity "getmelange.com/backend/models/identity"
	gdb "github.com/huntaub/go-db"

	adErrors "airdispat.ch/errors"
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/server"
	"airdispat.ch/wire"

	"getmelange.com/backend/connect"
	"getmelange.com/backend/models/messages"
	"getmelange.com/backend/notifications"
	"getmelange.com/backend/packaging"
)

const (
	fetchPublicFrequency  = 5 * time.Minute
	fetchPrivateFrequency = 10 * time.Second

	logFetcherPriv = "[FETCHER-PRIV]"
	logFetcherPub  = "[FETCHER-PUB]"
)

// Fetcher coordinates the actual downloading of messages from the
// AirDispatch network.
type Fetcher struct {
	// Tickers that represent when the Fetcher should begin
	// checking for new messages.
	Public  *time.Ticker
	Private *time.Ticker

	Packager *packaging.Packager

	Watchers []chan *messages.JSONMessage

	Client *connect.Client

	Cache *Store

	Translator *messages.Translator

	Tables *db.Tables
	Store  gdb.Executor

	lastPublic  uint64
	lastPrivate uint64
	quit        chan chan bool

	Initial  bool
	fetching bool
}

// CreateFetcher will initialize a new fetcher object that will begin
// checking for new messages.
func CreateFetcher(client *connect.Client, cache *Store, tables *db.Tables, store gdb.Executor) *Fetcher {
	f := &Fetcher{
		Public:   time.NewTicker(fetchPublicFrequency),
		Private:  time.NewTicker(fetchPrivateFrequency),
		Watchers: make([]chan *messages.JSONMessage, 0),
		quit:     make(chan chan bool),

		// Load in the initial information
		Client: client,
		Cache:  cache,
		Tables: tables,
		Store:  store,

		// Translator
		Translator: messages.CreateTranslator(tables, store),
	}

	go f.Start()
	return f
}

// startFetch will check for the appropriate updates as necessary.
func (f *Fetcher) Start() {
	if f.fetching {
		return
	}
	f.fetching = true

	for {
		select {
		case <-f.Public.C:
			if f.Client == nil {
				continue
			}

			go f.fetchPublic(f.lastPublic)
		case <-f.Private.C:
			if f.Client == nil {
				continue
			}

			go f.fetchPrivate(f.lastPrivate)
		case c := <-f.quit:
			c <- true
			return
		}
	}
}

func (f *Fetcher) getProfile(to string) (*messages.JSONProfile, error) {
	profile, found := f.Cache.RetrieveProfile(to)
	if !found {
		return f.addProfile(to, "")
	}

	return profile, nil
}

func (f *Fetcher) buildProfiles(addresses ...*identity.Address) map[string]*messages.JSONProfile {
	profiles := make(map[string]*messages.JSONProfile)
	for _, addr := range addresses {
		profile, err := f.getProfile(addr.Alias)
		if err != nil {
			logError("[FETCH-PROF]", "Receieved error getting profile.", err)
			continue
		}

		profiles[profile.Alias] = profile
		profiles[profile.Fingerprint] = profile
	}

	return profiles
}

func (f *Fetcher) translateMessage(msg *message.Mail, public bool, name string, context map[string][]byte) (*messages.JSONMessage, error) {
	header := msg.Header()

	// We need to create a profile object for all of the users.
	profiles := f.buildProfiles(append(header.To, header.From)...)

	return f.Translator.Request(&messages.TranslationRequest{
		Message:  msg,
		Public:   public,
		Name:     name,
		Context:  context,
		Profiles: profiles,
	})
}

func (f *Fetcher) fetchPrivate(since uint64) {
	// Update the time that we last looked for private messages.
	f.lastPrivate = uint64(time.Now().Unix())

	// Download all of the alerts that the Dispatcher has for us.
	messages, err := f.Client.DownloadMessages(since, true)
	if err != nil {
		logError(logFetcherPriv, "Received an error while downloading alerts from dispatcher.", err)
		return
	}

	if len(messages) > 0 {
		fmt.Println(logFetcherPriv, "Received", len(messages), "new private messages.")
	}

	// Range over the messages returned.
	for _, v := range messages {
		data, typ, h, err := v.Message.Reconstruct(f.Client.Origin, false)
		if err != nil || typ != wire.MessageDescriptionCode {
			logError(logFetcherPriv, "Received an error while reconstructing an alert.", err)
			continue
		}

		desc, err := server.CreateMessageDescriptionFromBytes(data, h)
		if err != nil {
			logError(logFetcherPriv, "Received an error while deserializing message description.", err)
			continue
		}

		json, err := f.addMessage(desc.Name, h.From.String(), desc.Location, v.Context)
		if err != nil {
			logError(logFetcherPriv, "Received an error while downloading a message.", err)
		}

		// Only Display Notifications if we have the ability
		// and if this is not the initial fetch.
		if f.Packager != nil && since > 0 {
			msg, err := notifications.CheckMessageForNotification(f.Packager, json)
			if err != nil {
				logError(logFetcherPriv, "Received an error requesting notification.", err)
			} else if msg != nil {
				// Display notification.
				msg.Display()
			}
		}
	}
}

// fetchProfileName is the location where profiles are stored on
// dispatcher servers.
const fetchProfileName = "profile"

func (f *Fetcher) addProfile(from, location string) (*messages.JSONProfile, error) {
	fetch := func() (*messages.JSONProfile, error) {
		profile, err := f.Client.GetMessage(fetchProfileName, from, location)
		if adErr, ok := err.(*adErrors.Error); ok {

			// We were unable to find a profile and must
			// contend with the default profile.
			if adErr.Code == 5 {
				return f.Translator.FromProfile(from, nil), nil
			}
		}

		if err != nil {
			return nil, err
		}

		return f.Translator.FromProfile(from, profile), nil
	}

	profile, err := fetch()
	if err != nil {
		return nil, err
	}

	f.Cache.AddProfile(profile, fetch)

	return profile, nil
}

func (f *Fetcher) addMessage(name, from, location string, context map[string][]byte) (*messages.JSONMessage, error) {
	// Construct the function that will check for updates on this message.
	fetch := func() (*messages.JSONMessage, error) {
		return f.getMessage(name, from, location, context)
	}

	// Actually download the message.
	json, err := fetch()
	if err != nil {
		return nil, err
	}

	// Alert the store that there is a new message.
	f.Cache.AddMessage(json, fetch)

	// Let the observers know that we have a new message.
	f.notifyObservers(json)

	return json, nil
}

// fetchPublic will attempt to check for new public messages from
// users we are following.
func (f *Fetcher) fetchPublic(since uint64) {
	// Update the last time that the messages were fetched.
	f.lastPublic = uint64(time.Now().Unix())

	// Get the addresses that we are subscribed to in the
	// database.
	var subscribed []*mIdentity.Address
	err := f.Tables.Address.Get().Where("subscribed", true).All(f.Store, &subscribed)
	if err != nil {
		logError(logFetcherPub, "Received error getting following users.", err)
		return
	}

	for _, user := range subscribed {
		_, err := f.addPublic(since, user)
		if err != nil {
			logError(logFetcherPub, "Received error downloading messages from user.", err)
		}
	}
}

func (f *Fetcher) addPublic(since uint64, user *mIdentity.Address) ([]*messages.JSONMessage, error) {
	// Download all of the messages associated with a
	// particular following user.
	msgs, err := f.getPublicMessages(since, user)
	if err != nil {
		return nil, err
	}

	for _, m := range msgs {
		f.notifyObservers(m)

		fetch := func() (*messages.JSONMessage, error) {
			if !strings.HasPrefix(m.Name, "__") {
				return f.getMessage(m.Name, m.From.Fingerprint, "", nil)
			}

			return nil, errors.New("Unable to update legacy message without name support.")
		}

		// Alert the store that there is a new message.
		f.Cache.AddMessage(m, fetch)
	}

	return msgs, nil
}

// getMessage will use the *connect.Client in order to download the
// message from the dispatcher.
func (f *Fetcher) getMessage(name, from, location string, context map[string][]byte) (*messages.JSONMessage, error) {
	// TODO: h.From _MUST_ be the server key, not the client key.
	mail, err := f.Client.GetMessage(name, from, location)
	if err != nil {
		return nil, err
	}

	// Translate the message into JSON.
	return f.translateMessage(mail, false, name, context)
}

// getPublicMessages will download all of the public messages
// published since `since` by author `addr`.
func (f *Fetcher) getPublicMessages(since uint64, addr *mIdentity.Address) ([]*messages.JSONMessage, error) {
	msg, err := f.Client.GetPublicMessages(since, addr)
	if adErr, ok := err.(*adErrors.Error); ok {
		// An adError is returned if there is no public mail.
		if adErr.Code == 5 {
			return []*messages.JSONMessage{}, nil
		}
	}

	// Report other types of errors as normal.
	if err != nil {
		return nil, err
	}

	output := make([]*messages.JSONMessage, len(msg))
	for i, v := range msg {
		output[i], err = f.translateMessage(v, true, "", nil)
		if err != nil {
			logError(logFetcherPriv, "Received message translating public message.", err)
		}
	}

	return output, nil
}

// CheckForMessages will run a public and private fetch right now
// given a manually overridden `since` value.
func (f *Fetcher) CheckForMessages(since uint64) {
	f.Initial = true
	f.fetchPublic(since)
	f.fetchPrivate(since)
}

// Stop will stop the fetcher from automatically looking for new
// messages. It must be restarted using the Start() function.
func (f *Fetcher) Stop() {
	if !f.fetching {
		return
	}
	f.fetching = false
	f.Initial = false

	c := make(chan bool)
	f.quit <- c
	<-c
}

// notifyObservers will send the JSONMessage to each channel that
// requested it.
func (f *Fetcher) notifyObservers(j *messages.JSONMessage) {
	for _, v := range f.Watchers {
		v <- j
	}
}

// AddObserver will add a channel that will receive the updates about
// new messages. This is generally used for websocket connections to
// the server.
func (f *Fetcher) AddObserver(j chan *messages.JSONMessage) {
	f.Watchers = append(f.Watchers, j)
}
