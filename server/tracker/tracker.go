package main

import (
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/tracker"
	"encoding/gob"
	"flag"
	"fmt"
	"os"
)

var port = flag.String("port", "2048", "select the port on which to run the tracking server")
var key_file = flag.String("key", "", "the file that will save or load your keys")
var save_file = flag.String("save", "", "the file that will save or load the tracker addresses")

var storedAddresses map[string]*message.SignedMessage
var aliasedAddresses map[string]*message.SignedMessage

func main() {
	flag.Parse()

	// Initialize the Database of Addresses
	storedAddresses = make(map[string]*message.SignedMessage)
	aliasedAddresses = make(map[string]*message.SignedMessage)

	loadedKey, err := identity.LoadKeyFromFile(*key_file)

	if err != nil {

		loadedKey, err = identity.CreateIdentity()
		if err != nil {
			fmt.Println("Unable to Create Tracker Key")
			return
		}

		if *key_file != "" {

			err = loadedKey.SaveKeyToFile(*key_file)
			if err != nil {
				fmt.Println("Unable to Save Tracker Key")
				return
			}
		}

	}
	fmt.Println("Loaded Address", loadedKey.Address.String())

	delegate := &myTracker{}
	theTracker := &tracker.Tracker{
		Key:      loadedKey,
		Delegate: delegate,
	}
	LoadRecords(delegate)
	theTracker.StartServer(*port)
}

func LoadRecords(t *myTracker) {
	if *save_file != "" {
		f, err := os.Open(*save_file)
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

		storedAddresses = out[0]
		aliasedAddresses = out[1]
	}
}

func SaveRecords(t *myTracker) {
	if *save_file != "" {
		f, err := os.Create(*save_file)
		if err != nil {
			t.LogMessage("Unable to create file:", err.Error())
			return
		}
		enc := gob.NewEncoder(f)
		err = enc.Encode([]map[string]*message.SignedMessage{storedAddresses, aliasedAddresses})
		if err != nil {
			t.LogMessage("Unable to encode messages:", err.Error())
			return
		}
	}
}

type myTracker struct {
	tracker.BasicTracker
}

func (t *myTracker) SaveRecord(address *identity.Address, record *message.SignedMessage, alias string) {
	// Store the RegisterdAddress in the Database
	storedAddresses[address.String()] = record

	if alias != "" {
		aliasedAddresses[alias] = record
	}
	go SaveRecords(t)
}

func (*myTracker) GetRecordByAddress(address *identity.Address) *message.SignedMessage {
	// Lookup the Address (by address) in the Database
	info, _ := storedAddresses[address.String()]
	return info
}

func (*myTracker) GetRecordByAlias(alias string) *message.SignedMessage {
	// Lookup the Address (by address) in the Database
	info, _ := aliasedAddresses[alias]
	return info
}
