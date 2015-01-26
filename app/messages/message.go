package messages

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	adErrors "airdispat.ch/errors"
	"airdispat.ch/routing"
	"airdispat.ch/server"
	"airdispat.ch/wire"
	"getmelange.com/app/models"
	"getmelange.com/app/notifications"
	"getmelange.com/app/packaging"
	"getmelange.com/dap"
	gdb "github.com/huntaub/go-db"
)

var (
	majorStore   = NewMessageStore()
	majorFetcher = NewFetcher()
)

const (
	updateFrequency        = 2 * time.Minute
	fetchPublicFrequency   = 5 * time.Minute
	fetchPrivateFrequency  = 10 * time.Second
	profileExpiryFrequency = 5 * time.Minute
)

type updateFunc func() (*models.JSONMessage, error)
type profileFunc func() (*models.JSONProfile, error)

type MessageStoreRecord struct {
	*models.JSONMessage
	Retrieved time.Time
	Expires   time.Time
	Public    bool
	Updater   updateFunc
}

type MessageStoreProfile struct {
	*models.JSONProfile
	Expires time.Time
	Updater profileFunc
}

type MessageStore struct {
	// Store Everything, only loop over this one
	Named map[string]*MessageStoreRecord
	// Quick lookup to Public Messages
	Public map[string][]*MessageStoreRecord
	// Quick lookup to Public Messages
	Profiles map[string]*MessageStoreProfile
	// Ticker
	Ticker *time.Ticker
	// Meta
	Lock        *sync.RWMutex
	ProfileLock *sync.RWMutex
	Latest      time.Time
}

func NewMessageStore() *MessageStore {
	c := &MessageStore{
		Named:       make(map[string]*MessageStoreRecord),
		Profiles:    make(map[string]*MessageStoreProfile),
		Public:      make(map[string][]*MessageStoreRecord),
		Ticker:      time.NewTicker(updateFrequency),
		Lock:        &sync.RWMutex{},
		ProfileLock: &sync.RWMutex{},
	}
	go c.updater()
	return c
}

func (c *MessageStore) updater() {
	for {
		now := <-c.Ticker.C
		fmt.Println("Updater is Running", now)

		c.Lock.Lock()
		now = time.Now()
		for key, v := range c.Named {
			if time.Now().After(v.Expires) {
				test := time.Now()
				fmt.Println("[CHECKING]", "about to update a message", key)
				msg, err := v.Updater()
				fmt.Println("[CHECKING]", time.Now().Sub(test), "to update a message")
				if err != nil {
					fmt.Println("Uh oh, got an error updating a message", err)
				} else {
					v.JSONMessage = msg
					v.Expires = c.getExpiryFromMessage(msg)
				}
			}
		}
		c.Lock.Unlock()

		c.ProfileLock.Lock()
		for _, v := range c.Profiles {
			if time.Now().After(v.Expires) {
				msg, err := v.Updater()
				if err != nil {
					fmt.Println("Uh oh, got an error updating a message", err)
				} else {
					v.JSONProfile = msg
					v.Expires = time.Now().Add(profileExpiryFrequency)
				}
			}
		}
		c.ProfileLock.Unlock()
	}
}

func (c *MessageStore) getExpiryFromMessage(msg *models.JSONMessage) time.Time {
	diff := float64((time.Now().Sub(msg.Date) / time.Minute) + 1)

	// 4.46489 log(1.56508 x)
	expiresIn := time.Duration(4.46489*math.Log(1.56508*diff)) * time.Minute

	return time.Now().Add(expiresIn)
}

func (c *MessageStore) AddMessage(msg *models.JSONMessage, refresh updateFunc) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	messageFrom := msg.From.Fingerprint
	address := fmt.Sprintf("%s/%s", messageFrom, msg.Name)

	expiry := c.getExpiryFromMessage(msg)

	retrieved := time.Now()

	record := &MessageStoreRecord{
		JSONMessage: msg,
		Retrieved:   retrieved,
		Expires:     expiry,
		Public:      msg.Public,
		Updater:     refresh,
	}

	// Update Latest
	if retrieved.After(c.Latest) {
		c.Latest = retrieved
	}

	c.Named[address] = record

	if msg.Public {
		if _, ok := c.Public[messageFrom]; !ok {
			c.Public[messageFrom] = []*MessageStoreRecord{
				record,
			}
		} else {
			c.Public[messageFrom] = append(c.Public[messageFrom], record)
		}
	}
}

func (c *MessageStore) AddProfile(msg *models.JSONProfile, refresh profileFunc) {
	c.ProfileLock.Lock()
	defer c.ProfileLock.Unlock()

	addr := msg.Fingerprint
	if msg.Fingerprint == "" {
		addr = msg.Alias
	}

	c.Profiles[addr] = &MessageStoreProfile{
		JSONProfile: msg,
		Expires:     time.Now().Add(profileExpiryFrequency),
		Updater:     refresh,
	}
}

func (c *MessageStore) RetrieveProfile(forAddr string) (*models.JSONProfile, bool) {
	c.ProfileLock.RLock()
	defer c.ProfileLock.RUnlock()

	if msg, ok := c.Profiles[forAddr]; ok {
		return msg.JSONProfile, true
	}
	return nil, false
}

func (c *MessageStore) RetrieveAll() []*models.JSONMessage {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	outputMessages := make([]*models.JSONMessage, len(c.Named))
	i := 0
	for _, v := range c.Named {
		outputMessages[i] = v.JSONMessage
		i++
	}

	return outputMessages
}

func (c *MessageStore) RetrieveMessage(from string, name string) *models.JSONMessage {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	address := fmt.Sprintf("%s/%s", from, name)

	if msg, ok := c.Named[address]; !ok {
		return nil
	} else {
		return msg.JSONMessage
	}
}

func (c *MessageStore) RetrieveAllPrivate() []*models.JSONMessage {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	outputMessages := make([]*models.JSONMessage, 0)

	for _, v := range c.Named {
		if !v.Public {
			outputMessages = append(outputMessages, v.JSONMessage)
		}
	}

	return outputMessages
}

func (c *MessageStore) RetrieveAllPublic() []*models.JSONMessage {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	outputMessages := make([]*models.JSONMessage, 0)

	for _, publicStore := range c.Public {
		for _, v := range publicStore {
			outputMessages = append(outputMessages, v.JSONMessage)
		}
	}

	return outputMessages
}

func (c *MessageStore) RetrievePublic(from string) []*models.JSONMessage {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	if msgs, ok := c.Public[from]; !ok {
		return nil
	} else {
		outputMessages := make([]*models.JSONMessage, len(msgs))
		for i, v := range msgs {
			outputMessages[i] = v.JSONMessage
		}
		return outputMessages
	}
}

func (c *MessageStore) LatestMessage() time.Time {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	return c.Latest
}

func (c *MessageStore) Invalidate() {
	c.Lock.Lock()
	c.Named = make(map[string]*MessageStoreRecord)
	c.Public = make(map[string][]*MessageStoreRecord)
	c.Lock.Unlock()

	c.ProfileLock.Lock()
	c.Profiles = make(map[string]*MessageStoreProfile)
	c.ProfileLock.Unlock()
}

type Fetcher struct {
	Public  *time.Ticker
	Private *time.Ticker

	Packager *packaging.Packager

	Watchers []chan *models.JSONMessage

	Client *dap.Client
	Router routing.Router

	Tables map[string]gdb.Table
	Store  gdb.Executor

	lastPublic  uint64
	lastPrivate uint64
	quit        chan chan bool
}

func NewFetcher() *Fetcher {
	f := &Fetcher{
		Public:   time.NewTicker(fetchPublicFrequency),
		Private:  time.NewTicker(fetchPrivateFrequency),
		Watchers: make([]chan *models.JSONMessage, 0),
		quit:     make(chan chan bool),
	}

	go f.startFetch()
	return f
}

func (f *Fetcher) startFetch() {
	for {
		select {
		case now := <-f.Public.C:
			if f.Client == nil {
				continue
			}

			fmt.Println("Fetching Public Messages", now)
			go f.fetchPublic(f.lastPublic)
		case now := <-f.Private.C:
			if f.Client == nil {
				continue
			}

			fmt.Println("Fetching Private Messages", now)
			go f.fetchPrivate(f.lastPrivate)
		case c := <-f.quit:
			c <- true
			return
		}
	}
}

func (f *Fetcher) fetchSpecificMessage(name, from, location string, context map[string][]byte) (*models.JSONMessage, error) {
	// TODO: h.From _MUST_ be the server key, not the client key.
	test := time.Now()
	mail, realErr := downloadMessage(f.Router, name, f.Client.Key, from, location)
	if realErr != nil {
		return nil, realErr
	}
	downloadDuration := time.Now().Sub(test)

	test = time.Now()
	output := translateMessageWithContext(f.Router, f.Client.Key, false, context, mail)
	if len(output) != 1 {
		return nil, errors.New("Translated to many messages")
	}
	translateDuration := time.Now().Sub(test)

	fmt.Println("Fetching Message Timing", downloadDuration, translateDuration)

	output[0].Name = name

	return output[0], nil
}

func (f *Fetcher) fetchPrivate(since uint64) {
	f.lastPrivate = uint64(time.Now().Unix())
	// Download Alerts

	messages, realErr := f.Client.DownloadMessages(since, true)
	if realErr != nil {
		fmt.Println("Got an error downloading messages", realErr)
		return
	}

	// Get Messages from Alerts
	for _, v := range messages {
		data, typ, h, realErr := v.Message.Reconstruct(f.Client.Key, false)
		if realErr != nil || typ != wire.MessageDescriptionCode {
			continue
		}

		desc, realErr := server.CreateMessageDescriptionFromBytes(data, h)
		if realErr != nil {
			fmt.Println("Can't create message description", realErr)
			continue
		}

		fetch := func() (*models.JSONMessage, error) {
			return f.fetchSpecificMessage(desc.Name, h.From.String(), desc.Location, v.Context)
		}
		json, realErr := fetch()

		if realErr == nil {
			majorStore.AddMessage(json, fetch)

			if f.Packager != nil {
				msg, err := notifications.CheckMessageForNotification(f.Packager, json)
				if err != nil {
					fmt.Println("Error checking for message.", err)
				} else if msg != nil {
					msg.Display()
				}
			}

			f.sendToWatchers(json)
		} else {
			fmt.Println("=== Error occurred getting message", realErr)
		}
	}
}

func (f *Fetcher) fetchSpecificPublic(since uint64, addr *models.Address) ([]*models.JSONMessage, error) {
	msg, realErr := downloadPublicMail(f.Router, since, f.Client.Key, addr)
	if realErr != nil {
		switch t := realErr.(type) {
		case *adErrors.Error:
			if t.Code == 5 {
				// No public mail at that address.
				return []*models.JSONMessage{}, nil
			} else {
				return nil, realErr
			}
		case error:
			return nil, realErr
		}
	}

	return translateMessage(f.Router, f.Client.Key, true, msg...), nil
}

func (f *Fetcher) fetchPublic(since uint64) {
	f.lastPublic = uint64(time.Now().Unix())

	var s []*models.Address
	realErr := f.Tables["address"].Get().Where("subscribed", true).All(f.Store, &s)
	if realErr != nil {
		fmt.Println("Couldn't get addresses subscribed to", realErr)
		return
	}

	for _, v := range s {
		msgs, err := f.fetchSpecificPublic(since, v)
		if err != nil {
			fmt.Println("Couldn't get specific public messages", err)
			continue
		}

		for _, m := range msgs {
			f.sendToWatchers(m)
			majorStore.AddMessage(m, func() (*models.JSONMessage, error) {
				if !strings.HasPrefix(m.Name, "__") {
					fmt.Println("Fetching (Public) Message...", m.Name, m.From.Fingerprint)
					return f.fetchSpecificMessage(m.Name, m.From.Fingerprint, "", nil)
				}
				return nil, errors.New("Unable to update a message without name support.")
			})
		}
	}

}

func (f *Fetcher) Stop() {
	c := make(chan bool)
	f.quit <- c
	<-c
}

func (f *Fetcher) ForceFetch(since uint64) {
	f.fetchPublic(since)
	f.fetchPrivate(since)
}

func (f *Fetcher) LoadCredentials(c *dap.Client, r routing.Router, t map[string]gdb.Table, s gdb.Executor) {
	// f.Stop()
	f.Client = c
	f.Router = r
	f.Tables = t
	f.Store = s
	go f.startFetch()
}

func (f *Fetcher) NeedsCredentials() bool {
	return f.Client == nil
}

func (f *Fetcher) sendToWatchers(j *models.JSONMessage) {
	for _, v := range f.Watchers {
		v <- j
	}
}

func (f *Fetcher) addWatcher(j chan *models.JSONMessage) {
	f.Watchers = append(f.Watchers, j)
}

func AddFetchWatcher(j chan *models.JSONMessage) {
	majorFetcher.addWatcher(j)
	return
}

func InvalidateCaches() {
	fmt.Println("Invalidating Caches")
	majorStore.Invalidate()
	majorFetcher.Stop()
	majorFetcher.Client = nil
}

// CreateMessage(name, to, from, alert, component...)
// m.UpdateMessage()    [ Client -> Server ]
// m.RetrieveMessage()  [ Client <- Server ]
// m.Delete()

type MessageManager struct {
	Tables   map[string]gdb.Table
	Store    gdb.Executor
	Client   *dap.Client
	Router   routing.Router
	Identity *models.Identity
}

func (m *MessageManager) PublishMessage(msg *models.JSONMessage) error {
	mail, to, err := msg.ToDispatch(m.Client.Key)
	if err != nil {
		return err
	}

	// Add to Database
	modelMsg, modelComp := msg.ToModel(m.Client.Key)
	_, err = m.Tables["message"].Insert(modelMsg).Exec(m.Store)
	if err != nil {
		return err
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
	name, err := m.Client.PublishMessage(mail, to, msg.Name, !msg.Public)
	if err != nil {
		return err
	}

	if !msg.Public {
		fmt.Println("Sending alerts...")
		// Send Alert
		var errs []error

		for _, v := range msg.To {
			fmt.Println("Sending alert to", v.Alias)
			err = SendAlert(m.Router, name, m.Client.Key, v.Alias, m.Client.Server.Alias)
			if err != nil {
				fmt.Println("Got error sending alert", err)
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			return errs[0]
		}
	}

	// Load the new message into the webpage.
	go func() {
		msg.Self = true

		myProfile, myAlias, _ := m.currentProfile()

		msg.From = &models.JSONProfile{
			Name:        myProfile.Name,
			Avatar:      myProfile.Image,
			Alias:       myAlias.String(),
			Fingerprint: m.Client.Key.Address.String(),
		}

		majorFetcher.sendToWatchers(msg)
	}()

	return nil
}

func (m *MessageManager) GetMessage(alias string, name string) (*models.JSONMessage, error) {
	fmt.Println("[DEPRECEATED] Call to deprecated method GetMessage.")
	srv, author, err := GetAddresses(m.Router, &models.Address{
		Alias: alias,
	})
	if err != nil {
		return nil, err
	}

	// TODO: h.From _MUST_ be the server key, not the client key.
	mail, err := downloadMessageFromServer(name, m.Client.Key, author, srv)
	if err != nil {
		return nil, err
	}

	jsonMsg := translateMessage(m.Router, m.Client.Key, false, mail)
	if len(jsonMsg) != 1 {
		return nil, errors.New("models.JSONMessage translated too many messages.")
	}

	return jsonMsg[0], nil
}

func (m *MessageManager) GetPublic(since uint64, addr *models.Address) ([]*models.JSONMessage, error) {
	fmt.Println("[DEPRECEATED] Call to deprecated method GetPublic.")
	msg, realErr := downloadPublicMail(m.Router, since, m.Client.Key, addr)
	if realErr != nil {
		switch t := realErr.(type) {
		case *adErrors.Error:
			if t.Code == 5 {
				// No public mail at that address.
				return []*models.JSONMessage{}, nil
			} else {
				return nil, realErr
			}
		case error:
			return nil, realErr
		}
	}

	return translateMessage(m.Router, m.Client.Key, true, msg...), nil
}

func (m *MessageManager) currentProfile() (*models.Profile, *models.Alias, error) {
	myAlias := &models.Alias{}
	err := m.Tables["alias"].Get().Where("identity", m.Identity.Id).One(m.Store, myAlias)
	if err != nil {
		return nil, nil, err
	}

	myProfile := &models.Profile{}
	err = m.Identity.Profile.One(m.Store, myProfile)
	if err != nil || myProfile.Id == 0 {
		fmt.Println("Unable to get my profile.", err)
		myProfile = &models.Profile{
			Name:  myAlias.String(),
			Image: defaultProfileImage(m.Client.Key.Address),
		}
	}

	if strings.Contains(myProfile.Image, "@") {
		myProfile.Image = fmt.Sprintf("http://data.melange:7776/%s", myProfile.Image)
	}

	return myProfile, myAlias, nil
}

func (m *MessageManager) GetSentMessages(since uint64, withComponents []string) ([]*models.JSONMessage, error) {
	fmt.Println("[DEPRECEATED] Call to deprecated method GetSentMessages(?).")
	msgs := make([]*models.Message, 0)
	realErr := m.Tables["message"].Get().Where("from", m.Client.Key.Address.String()).All(m.Store, &msgs)
	if realErr != nil {
		return nil, realErr
	}

	myProfile, myAlias, err := m.currentProfile()
	if err != nil {
		return nil, err
	}

	var outputMessages []*models.JSONMessage
	for _, v := range msgs {
		msg, err := translateModel(m.Router, m.Tables["component"], m.Store, v, m.Client, myProfile, myAlias)
		if err != nil {
			fmt.Println("Couldn't translate model", err)
		}
		outputMessages = append(outputMessages, msg)
	}

	return outputMessages, nil
}

func (m *MessageManager) GetAllMessages(msgChan chan *models.JSONMessage) {
	var newData models.JSONMessageList
	if majorFetcher.NeedsCredentials() {
		fmt.Println("Loading Credentials through Force Fetch")
		m.forceFetch()
	} else {
		fmt.Println("Credentials already loaded.")
		newData = models.JSONMessageList(append(majorStore.RetrieveAllPublic(), majorStore.RetrieveAllPrivate()...))
	}

	fmt.Println("Getting sent Messages")
	data, err := m.GetSentMessages(0, []string{})
	if err != nil {
		fmt.Println("Error getting Sent Messages", err)
	}
	newData = append(newData, data...)

	sort.Reverse(newData)

	for _, v := range newData {
		msgChan <- v
	}
}

func (m *MessageManager) forceFetch() {
	fmt.Println("Loading Credentials")
	majorFetcher.LoadCredentials(m.Client, m.Router, m.Tables, m.Store)
	fmt.Println("Forcing Fetch")
	majorFetcher.ForceFetch(0)
}

func (m *MessageManager) GetPublicMessages(since uint64, withComponents []string) []*models.JSONMessage {
	fmt.Println("[DEPRECEATED] Call to deprecated method GetPublicMessages.")
	if majorFetcher.NeedsCredentials() {
		go m.forceFetch()
	}

	return majorStore.RetrieveAllPublic()
}

// GetPrivateMessages will return an array of JSON Messges with the given arguments.
func (m *MessageManager) GetPrivateMessages(since uint64, withComponents []string) []*models.JSONMessage {
	fmt.Println("[DEPRECEATED] Call to deprecated method GetPrivateMessages.")
	if majorFetcher.NeedsCredentials() {
		go m.forceFetch()
	}

	return majorStore.RetrieveAllPrivate()
}
