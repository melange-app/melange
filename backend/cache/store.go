package cache

import (
	"fmt"
	"math"
	"sync"
	"time"

	"getmelange.com/backend/models/messages"
)

const (
	// updateFrequency is the frequency at which we check to see
	// if there are any messages that have updates.
	updateFrequency = 2 * time.Minute

	// profileExpiryFrequency is the frequency at which we check
	// to see if there are any profiles that have updates.
	profileExpiryFrequency = 5 * time.Minute
)

// These functions allow the objects to manually effect how the data
// is updated.
type updateFunc func() (*messages.JSONMessage, error)
type profileFunc func() (*messages.JSONProfile, error)

type record struct {
	Retrieved time.Time
	Expires   time.Time
}

// MessageRecord is the data that is contained in the cache about a
// specific AirDispatch record.
type MessageRecord struct {
	*messages.JSONMessage
	Public  bool
	Updater updateFunc

	record
}

// ProfileRecord is the data that is contained in the cache about a
// specific Melange profile.
type ProfileRecord struct {
	*messages.JSONProfile
	Updater profileFunc

	record
}

// Store is an object that keeps track of the messages that we are
// caching.
type Store struct {
	// Store Everything, only loop over this one
	Named map[string]*MessageRecord

	// Quick lookup to Public Messages
	Public map[string][]*MessageRecord

	// Quick lookup to Public Messages
	Profiles map[string]*ProfileRecord

	// Ticker
	Ticker *time.Ticker

	// Meta
	Lock        *sync.RWMutex
	ProfileLock *sync.RWMutex
	Latest      time.Time
}

// CreateStore will initialize a new Store object that starts it's
// updating timer immediately.
func CreateStore() *Store {
	c := &Store{
		Named:       make(map[string]*MessageRecord),
		Profiles:    make(map[string]*ProfileRecord),
		Public:      make(map[string][]*MessageRecord),
		Ticker:      time.NewTicker(updateFrequency),
		Lock:        &sync.RWMutex{},
		ProfileLock: &sync.RWMutex{},
	}

	go func(c *Store) {
		// Loop every `updateFrequency`, checking for new
		// messages.
		for {
			<-c.Ticker.C
			c.checkMessageUpdates()
			c.checkProfileUpdates()
		}
	}(c)

	return c
}

// checkMessageUpdates will see if any of the stored Messages require
// an update at this time.
func (c *Store) checkMessageUpdates() {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	// Loop through each of the messages that we have stored.
	for key, v := range c.Named {

		// Run the updater if the cache has expired.
		if time.Now().After(v.Expires) {
			msg, err := v.Updater()

			if err != nil {
				fmt.Printf("Error occurred while updating message named %s.\n", key)
				fmt.Println(err)
				continue
			}

			// Update the message inside the value.
			v.JSONMessage = msg
			v.Expires = c.getMessageExpiration(msg)
		}
	}
}

// checkProfileUpdates will see if any of the stored Profiles require
// an update at this time.
func (c *Store) checkProfileUpdates() {
	c.ProfileLock.Lock()
	defer c.ProfileLock.Unlock()

	for _, v := range c.Profiles {

		if time.Now().After(v.Expires) {
			msg, err := v.Updater()

			if err != nil {
				fmt.Println("Error occurred while updating profile.")
				fmt.Println(err)
				continue
			}

			v.JSONProfile = msg
			v.Expires = time.Now().Add(profileExpiryFrequency)
		}
	}
}

// getMessageExpiration will take a message that has been updated and
// calculate when we should next check for updates.
// TODO: Move to the JSONMessage model.
func (c *Store) getMessageExpiration(msg *messages.JSONMessage) time.Time {
	diff := float64((time.Now().Sub(msg.Date) / time.Minute) + 1)

	// 4.46489 log(1.56508 x)
	expiresIn := time.Duration(4.46489*math.Log(1.56508*diff)) * time.Minute

	return time.Now().Add(expiresIn)
}

// AddMessage will add a JSONMessage into the cache.
func (c *Store) AddMessage(msg *messages.JSONMessage, refresh updateFunc) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	messageFrom := msg.From.Fingerprint
	address := fmt.Sprintf("%s/%s", messageFrom, msg.Name)

	expiry := c.getMessageExpiration(msg)

	retrieved := time.Now()

	record := &MessageRecord{
		JSONMessage: msg,
		Public:      msg.Public,
		Updater:     refresh,
		record: record{
			Retrieved: retrieved,
			Expires:   expiry,
		},
	}

	// Update Latest
	if retrieved.After(c.Latest) {
		c.Latest = retrieved
	}

	c.Named[address] = record

	// Add the message to the public list (for profile viewing) if
	// necessary.
	if msg.Public {
		c.Public[messageFrom] = append(c.Public[messageFrom], record)
	}
}

// AddProfile will add a new profile to the cached store.
func (c *Store) AddProfile(msg *messages.JSONProfile, refresh profileFunc) {
	c.ProfileLock.Lock()
	defer c.ProfileLock.Unlock()

	addr := msg.Fingerprint
	if msg.Fingerprint == "" {
		addr = msg.Alias
	}

	retrieved := time.Now()
	expiry := retrieved.Add(profileExpiryFrequency)

	c.Profiles[addr] = &ProfileRecord{
		JSONProfile: msg,
		Updater:     refresh,
		record: record{
			Retrieved: retrieved,
			Expires:   expiry,
		},
	}
}

// RetrieveProfile will get a profile for an address out of the
// store. It will return a boolean that represents whether or not it
// was found in the store.
func (c *Store) RetrieveProfile(forAddr string) (*messages.JSONProfile, bool) {
	c.ProfileLock.RLock()
	defer c.ProfileLock.RUnlock()

	if msg, ok := c.Profiles[forAddr]; ok {
		return msg.JSONProfile, true
	}
	return nil, false
}

// RetrieveMessage will get a message out of the store with a name and string. It will
func (c *Store) RetrieveMessage(from string, name string) (*messages.JSONMessage, bool) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	address := fmt.Sprintf("%s/%s", from, name)

	if msg, found := c.Named[address]; found {
		return msg.JSONMessage, true
	}

	return nil, false
}

// RetirevePublic will return all of the public messages cached from
// the user specified.
func (c *Store) RetrievePublic(from string) ([]*messages.JSONMessage, bool) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	msgs, found := c.Public[from]
	if !found {
		return nil, false
	}

	outputMessages := make([]*messages.JSONMessage, len(msgs))
	for i, v := range msgs {
		outputMessages[i] = v.JSONMessage
	}
	return outputMessages, true
}

// RetrieveAll will return all of the currently cached messages.
func (c *Store) RetrieveAll() []*messages.JSONMessage {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	outputMessages := make([]*messages.JSONMessage, len(c.Named))
	i := 0
	for _, v := range c.Named {
		outputMessages[i] = v.JSONMessage
		i++
	}

	return outputMessages
}

// RetrieveAllPrivate will return all of the currently cached messages
// sent directly to the user.
func (c *Store) RetrieveAllPrivate() []*messages.JSONMessage {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	var outputMessages []*messages.JSONMessage
	for _, v := range c.Named {
		if !v.Public {
			outputMessages = append(outputMessages, v.JSONMessage)
		}
	}

	return outputMessages
}

// RetrieveAllPublic will return all of the currently cached messages
// that are public.
func (c *Store) RetrieveAllPublic() []*messages.JSONMessage {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	var outputMessages []*messages.JSONMessage
	for _, publicStore := range c.Public {
		for _, v := range publicStore {
			outputMessages = append(outputMessages, v.JSONMessage)
		}
	}

	return outputMessages
}

// LatestMessage returns the timestamp of the most recent message that
// is located in the store.
func (c *Store) LatestMessage() time.Time {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	return c.Latest
}

// Clear will remove everything from the cache.
func (c *Store) Clear() {
	c.Lock.Lock()
	c.Named = make(map[string]*MessageRecord)
	c.Public = make(map[string][]*MessageRecord)
	c.Lock.Unlock()

	c.ProfileLock.Lock()
	c.Profiles = make(map[string]*ProfileRecord)
	c.ProfileLock.Unlock()
}

// TODO: Add load from disk functionality.
