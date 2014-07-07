package tracker

import (
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/tracker"
	"encoding/gob"
	"fmt"
	"os"
)

type Tracker struct {
	// FS Information
	KeyFile  string
	SaveFile string
	// Storage
	StoredAddresses  map[string]*message.SignedMessage
	AliasedAddresses map[string]*message.SignedMessage
	// Composition
	tracker.BasicTracker
}

func (t *Tracker) Run(port int) error {
	// Initialize the Database of Addresses
	t.StoredAddresses = make(map[string]*message.SignedMessage)
	t.AliasedAddresses = make(map[string]*message.SignedMessage)

	loadedKey, err := identity.LoadKeyFromFile(t.KeyFile)

	if err != nil {

		loadedKey, err = identity.CreateIdentity()
		if err != nil {
			return err
		}

		if t.KeyFile != "" {

			err = loadedKey.SaveKeyToFile(t.KeyFile)
			if err != nil {
				return err
			}
		}

	}
	fmt.Println("Loaded Address", loadedKey.Address.String())

	theTracker := &tracker.Tracker{
		Key:      loadedKey,
		Delegate: t,
	}
	t.LoadRecords()
	theTracker.StartServer(fmt.Sprintf("%d", port))
	return nil
}

func (t *Tracker) LoadRecords() {
	if t.SaveFile != "" {
		f, err := os.Open(t.SaveFile)
		if err != nil {
			t.LogMessage("Unable to open file:", err.Error())
			return
		}
		enc := gob.NewDecoder(f)
		var out []map[string]*message.SignedMessage
		err = enc.Decode(&out)
		if err != nil {
			t.LogMessage("Unable to decode messages:", err.Error())
			return
		}

		t.StoredAddresses = out[0]
		t.AliasedAddresses = out[1]
	}
}

func (t *Tracker) SaveRecords() {
	if t.SaveFile != "" {
		f, err := os.Create(t.SaveFile)
		if err != nil {
			t.LogMessage("Unable to create file:", err.Error())
			return
		}
		enc := gob.NewEncoder(f)
		err = enc.Encode([]map[string]*message.SignedMessage{t.StoredAddresses, t.AliasedAddresses})
		if err != nil {
			t.LogMessage("Unable to encode messages:", err.Error())
			return
		}
	}
}

func (t *Tracker) SaveRecord(address *identity.Address, record *message.SignedMessage, alias string) {
	// Store the RegisterdAddress in the Database
	t.StoredAddresses[address.String()] = record

	if alias != "" {
		t.AliasedAddresses[alias] = record
	}
	go t.SaveRecords()
}

func (t *Tracker) GetRecordByAddress(address *identity.Address) *message.SignedMessage {
	// Lookup the Address (by address) in the Database
	info, _ := t.StoredAddresses[address.String()]
	return info
}

func (t *Tracker) GetRecordByAlias(alias string) *message.SignedMessage {
	// Lookup the Address (by address) in the Database
	info, _ := t.AliasedAddresses[alias]
	return info
}
