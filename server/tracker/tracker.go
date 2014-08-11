package tracker

import (
	"bytes"
	"encoding/hex"
	"fmt"

	gdb "github.com/huntaub/go-db"
	"github.com/jmoiron/sqlx"

	"encoding/gob"

	"airdispat.ch/crypto"
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/tracker"

	// Imported for DB Initialization
	_ "github.com/lib/pq"
)

type Tracker struct {
	// FS Information
	KeyFile string
	// Database Info
	DBString string
	table    gdb.Table
	ex       *sqlx.DB
	// Composition
	tracker.BasicTracker
}

func (t *Tracker) Run(port int) error {
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
	fmt.Println("Loaded Encryption Key", hex.EncodeToString(crypto.RSAToBytes(loadedKey.Address.EncryptionKey)))

	theTracker := &tracker.Tracker{
		Key:      loadedKey,
		Delegate: t,
	}

	t.ex, err = sqlx.Open("postgres", t.DBString)
	if err != nil {
		return err
	}

	t.table, err = CreateTables(t.ex)
	if err != nil {
		return err
	}

	theTracker.StartServer(fmt.Sprintf("%d", port))
	return nil
}

func (t *Tracker) SaveRecord(address *identity.Address, record *message.SignedMessage, alias string) {
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)

	err := enc.Encode(record)
	if err != nil {
		fmt.Println("Error encoding", err)
		return
	}

	finding := &Record{}
	err = t.table.Get().Where("address", address.String()).One(t.ex, finding)
	if err != nil || finding.Address == "" {
		_, err = t.table.Insert(&Record{
			Address: address.String(),
			Alias:   alias,
			Message: data.Bytes(),
		}).Exec(t.ex)
		if err != nil {
			fmt.Println("Error inserting", err)
		}
	}

	finding.Alias = alias
	finding.Message = data.Bytes()

	_, err = t.table.Update(finding).Exec(t.ex)
	if err != nil {
		fmt.Println("Error updating", err)
	}
}

func (t *Tracker) GetRecordByAddress(address *identity.Address) *message.SignedMessage {
	return t.recordMessage("address", address.String())
}

func (t *Tracker) GetRecordByAlias(alias string) *message.SignedMessage {
	return t.recordMessage("alias", alias)
}

func (t *Tracker) recordMessage(col string, obj interface{}) *message.SignedMessage {
	record := &Record{}
	fmt.Println("Getting", col, obj)
	err := t.table.Get().Where(col, obj).One(t.ex, record)
	if err != nil || record.Address == "" {
		fmt.Println("Cannot get address", err)
		return nil
	}

	data := bytes.NewReader(record.Message)
	dec := gob.NewDecoder(data)

	m := &message.SignedMessage{}
	err = dec.Decode(&m)
	if err != nil {
		fmt.Println("Cannot decode SignedMessage", err)
		return nil
	}

	return m
}
