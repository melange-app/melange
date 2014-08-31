package models

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	adErrors "airdispat.ch/errors"
	"airdispat.ch/routing"
	"airdispat.ch/server"
	"airdispat.ch/wire"
	"getmelange.com/dap"
	gdb "github.com/huntaub/go-db"
)

var (
	majorStore   = NewMessageStore()
	majorFetcher = NewFetcher()
)

type updateFunc func() (*JSONMessage, error)
type profileFunc func() (*JSONProfile, error)

type MessageStoreRecord struct {
	*JSONMessage
	Retrieved time.Time
	Expires   time.Time
	Public    bool
	Updater   updateFunc
}

type MessageStoreProfile struct {
	*JSONProfile
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
	Lock   *sync.RWMutex
	Latest time.Time
}

func NewMessageStore() *MessageStore {
	c := &MessageStore{
		Named:    make(map[string]*MessageStoreRecord),
		Profiles: make(map[string]*MessageStoreProfile),
		Public:   make(map[string][]*MessageStoreRecord),
		Ticker:   time.NewTicker(2 * time.Minute),
		Lock:     &sync.RWMutex{},
	}
	go c.updater()
	return c
}

func (c *MessageStore) updater() {
	for {
		now := <-c.Ticker.C
		fmt.Println("Checking to update messages", now)

		c.Lock.Lock()
		for _, v := range c.Named {
			if time.Now().After(v.Expires) {
				msg, err := v.Updater()
				if err != nil {
					fmt.Println("Uh oh, got an error updating a message", err)
				} else {
					v.JSONMessage = msg
					v.Expires = c.getExpiryFromMessage(msg)
				}
			}
		}

		for _, v := range c.Profiles {
			if time.Now().After(v.Expires) {
				msg, err := v.Updater()
				if err != nil {
					fmt.Println("Uh oh, got an error updating a message", err)
				} else {
					v.JSONProfile = msg
					v.Expires = time.Now().Add(5 * time.Minute)
				}
			}
		}
		c.Lock.Unlock()
	}
}

func (c *MessageStore) getExpiryFromMessage(msg *JSONMessage) time.Time {
	diff := float64(time.Now().Sub(msg.Date))

	// 1.13277x10^11 e^(4.34626x10^-12 x)
	expiresIn := time.Duration(math.Exp(diff*4.34626e-12) * 1.13277e+11)

	return time.Now().Add(expiresIn)
}

func (c *MessageStore) AddMessage(msg *JSONMessage, refresh updateFunc) {
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

func (c *MessageStore) AddProfile(msg *JSONProfile, refresh profileFunc) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	addr := msg.Fingerprint
	if msg.Fingerprint == "" {
		addr = msg.Alias
	}

	c.Profiles[addr] = &MessageStoreProfile{
		JSONProfile: msg,
		Expires:     time.Now().Add(5 * time.Minute),
		Updater:     refresh,
	}
}

func (c *MessageStore) RetrieveProfile(forAddr string) (*JSONProfile, bool) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	if msg, ok := c.Profiles[forAddr]; ok {
		return msg.JSONProfile, true
	}
	return nil, false
}

func (c *MessageStore) RetrieveAll() []*JSONMessage {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	outputMessages := make([]*JSONMessage, len(c.Named))
	i := 0
	for _, v := range c.Named {
		outputMessages[i] = v.JSONMessage
		i++
	}

	return outputMessages
}

func (c *MessageStore) RetrieveMessage(from string, name string) *JSONMessage {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	address := fmt.Sprintf("%s/%s", from, name)

	if msg, ok := c.Named[address]; !ok {
		return nil
	} else {
		return msg.JSONMessage
	}
}

func (c *MessageStore) RetrieveAllPrivate() []*JSONMessage {
	fmt.Println("Retrieving Private")
	now := time.Now()
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	fmt.Println("Lock took", time.Now().Sub(now))
	now = time.Now()
	defer func() {
		fmt.Println("Retrieved Private, took", time.Now().Sub(now))
	}()

	outputMessages := make([]*JSONMessage, 0)

	for _, v := range c.Named {
		if !v.Public {
			outputMessages = append(outputMessages, v.JSONMessage)
		}
	}

	return outputMessages
}

func (c *MessageStore) RetrieveAllPublic() []*JSONMessage {
	fmt.Println("Retrieving Public")
	now := time.Now()
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	fmt.Println("Lock took", time.Now().Sub(now))
	now = time.Now()
	defer func() {
		fmt.Println("Retrieved Public, took", time.Now().Sub(now))
	}()

	outputMessages := make([]*JSONMessage, 0)

	for _, publicStore := range c.Public {
		for _, v := range publicStore {
			outputMessages = append(outputMessages, v.JSONMessage)
		}
	}

	return outputMessages
}

func (c *MessageStore) RetrievePublic(from string) []*JSONMessage {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	if msgs, ok := c.Public[from]; !ok {
		return nil
	} else {
		outputMessages := make([]*JSONMessage, len(msgs))
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
	c.Named = make(map[string]*MessageStoreRecord)
	c.Public = make(map[string][]*MessageStoreRecord)
}

type Fetcher struct {
	Public  *time.Ticker
	Private *time.Ticker

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
		Public:  time.NewTicker(5 * time.Minute),
		Private: time.NewTicker(10 * time.Second),
		quit:    make(chan chan bool),
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
			f.lastPublic = uint64(now.Unix())
		case now := <-f.Private.C:
			if f.Client == nil {
				continue
			}

			fmt.Println("Fetching Private Messages", now)
			go f.fetchPrivate(f.lastPrivate)
			f.lastPrivate = uint64(now.Unix())
		case c := <-f.quit:
			c <- true
			return
		}
	}
}

func (f *Fetcher) fetchSpecificMessage(name, from, location string, context map[string][]byte) (*JSONMessage, error) {
	// TODO: h.From _MUST_ be the server key, not the client key.
	mail, realErr := downloadMessage(f.Router, name, f.Client.Key, from, location)
	if realErr != nil {
		return nil, realErr
	}

	output := translateMessageWithContext(f.Router, f.Client.Key, false, context, mail)
	if len(output) != 1 {
		return nil, errors.New("Translated to many messages")
	}

	return output[0], nil
}

func (f *Fetcher) fetchPrivate(since uint64) {
	// Download Alerts
	fmt.Println("Fetching private since", since)

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

		now := time.Now()
		fetch := func() (*JSONMessage, error) {
			return f.fetchSpecificMessage(desc.Name, h.From.String(), desc.Location, v.Context)
		}
		json, realErr := fetch()
		if realErr == nil {
			majorStore.AddMessage(json, fetch)
		}
		fmt.Println("Added Message", time.Now().Sub(now))
	}
}

func (f *Fetcher) fetchSpecificPublic(since uint64, addr *Address) ([]*JSONMessage, error) {
	msg, realErr := downloadPublicMail(f.Router, since, f.Client.Key, addr)
	if realErr != nil {
		switch t := realErr.(type) {
		case *adErrors.Error:
			if t.Code == 5 {
				// No public mail at that address.
				return []*JSONMessage{}, nil
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
	fmt.Println("Fetching public since", since)

	var s []*Address
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
			majorStore.AddMessage(m, func() (*JSONMessage, error) {
				return f.fetchSpecificMessage(m.Name, m.From.Fingerprint, "", nil)
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
	fmt.Println("Forcing Fetch")
	f.fetchPublic(since)
	fmt.Println("Finished Public")
	f.fetchPrivate(since)
	fmt.Println("Finished Private")
}

func (f *Fetcher) LoadCredentials(c *dap.Client, r routing.Router, t map[string]gdb.Table, s gdb.Executor) {
	f.Stop()
	f.Client = c
	f.Router = r
	f.Tables = t
	f.Store = s
	go f.startFetch()
}

func (f *Fetcher) NeedsCredentials() bool {
	return f.Client == nil
}

// A full AirDispatch Message
type Message struct {
	Id gdb.PrimaryKey
	// Name registered on Server
	Name string
	// Cleared Addresses
	To string
	// Fingerprint of Sender
	From string
	// Message Metadata
	Alert    bool
	Incoming bool
	Date     int64

	Components *gdb.HasMany `table:"component" on:"message"`
}

func InvalidateCaches() {
	fmt.Println("Invalidating Caches")
	majorStore.Invalidate()
}

// CreateMessage(name, to, from, alert, component...)
// m.UpdateMessage()    [ Client -> Server ]
// m.RetrieveMessage()  [ Client <- Server ]
// m.Delete()

// An AirDispatch Component
type Component struct {
	Id      gdb.PrimaryKey
	Message *gdb.HasOne `table:"message"`
	Name    string
	Data    []byte
}

type MessageManager struct {
	Tables   map[string]gdb.Table
	Store    gdb.Executor
	Client   *dap.Client
	Router   routing.Router
	Identity *Identity
}

func (m *MessageManager) GetMessage(alias string, name string) (*JSONMessage, error) {
	srv, author, err := getAddresses(m.Router, &Address{
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
		return nil, errors.New("JSONMessage translated too many messages.")
	}

	return jsonMsg[0], nil
}

func (m *MessageManager) GetPublic(since uint64, addr *Address) ([]*JSONMessage, error) {
	msg, realErr := downloadPublicMail(m.Router, since, m.Client.Key, addr)
	if realErr != nil {
		switch t := realErr.(type) {
		case *adErrors.Error:
			if t.Code == 5 {
				// No public mail at that address.
				return []*JSONMessage{}, nil
			} else {
				return nil, realErr
			}
		case error:
			return nil, realErr
		}
	}

	return translateMessage(m.Router, m.Client.Key, true, msg...), nil
}

func (m *MessageManager) GetSentMessages(since uint64, withComponents []string) ([]*JSONMessage, error) {
	now := time.Now()
	defer func() {
		fmt.Println("Got sent, took", time.Now().Sub(now))
	}()

	msgs := make([]*Message, 0)
	realErr := m.Tables["message"].Get().Where("from", m.Client.Key.Address.String()).All(m.Store, &msgs)
	if realErr != nil {
		return nil, realErr
	}

	myAlias := &Alias{}
	realErr = m.Tables["alias"].Get().Where("identity", m.Identity.Id).One(m.Store, myAlias)
	if realErr != nil {
		return nil, realErr
	}

	myProfile := &Profile{}
	realErr = m.Identity.Profile.One(m.Store, myProfile)
	if realErr != nil || myProfile.Id == 0 {
		fmt.Println("Unable to get my profile.", realErr)
		myProfile = &Profile{
			Name:  myAlias.String(),
			Image: defaultProfileImage(m.Client.Key.Address),
		}
	}

	begin := time.Now()
	var outputMessages []*JSONMessage
	for _, v := range msgs {
		msg, err := translateModel(m.Router, m.Tables["component"], m.Store, v, m.Client, myProfile, myAlias)
		if err != nil {
			fmt.Println("Couldn't translate model", err)
		}
		outputMessages = append(outputMessages, msg)
	}
	fmt.Println("Intermediate", time.Now().Sub(begin))

	return outputMessages, nil
}

func (m *MessageManager) GetPublicMessages(since uint64, withComponents []string) []*JSONMessage {
	if majorFetcher.NeedsCredentials() {
		fmt.Println("Forcing Fetch")
		majorFetcher.LoadCredentials(m.Client, m.Router, m.Tables, m.Store)
		majorFetcher.ForceFetch(0)
	}
	fmt.Println("Getting Public Messages")
	return majorStore.RetrieveAllPublic()
}

// GetPrivateMessages will return an array of JSON Messges with the given arguments.
func (m *MessageManager) GetPrivateMessages(since uint64, withComponents []string) []*JSONMessage {
	if majorFetcher.NeedsCredentials() {
		fmt.Println("Forcing Fetch")
		majorFetcher.LoadCredentials(m.Client, m.Router, m.Tables, m.Store)
		majorFetcher.ForceFetch(0)
	}
	fmt.Println("Getting Private Messages")
	return majorStore.RetrieveAllPrivate()
}
