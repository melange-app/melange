package dispatcher

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"melange/dap"

	"airdispat.ch/crypto"
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/server"
	"airdispat.ch/tracker"
	"github.com/coopernurse/gorp"

	// Imported for DB Initialization
	_ "github.com/lib/pq"
)

func createError(location string, error error) *server.ServerError {
	return &server.ServerError{
		Location: location,
		Error:    error,
	}
}

// Server exposes a structure that holds all the properties of an AirDisaptch
// dispatcher.
type Server struct {
	// Basic Properties
	Me      string
	KeyFile string
	Key     *identity.Identity

	// Tracker Properties
	TrackerURL string
	Alias      string

	// Database
	DBConn string
	DBType string
	dbmap  *gorp.DbMap

	// Import Logging and Such
	server.BasicServer
}

// Run will start the server with the specified database model.
func (m *Server) Run(port int) error {
	// Initialize Type of Database
	var dialect gorp.Dialect
	switch m.DBType {
	case "postgres":
		dialect = gorp.PostgresDialect{}
	case "sqlite":
		dialect = gorp.SqliteDialect{}
	case "mysql":
		dialect = gorp.MySQLDialect{}
	case "mssql":
		dialect = gorp.SqlServerDialect{}
	default:
		return errors.New("Unable to determine DB dialect.")
	}

	// Initialize Database Connection
	db, err := sql.Open(m.DBType, m.DBConn)
	if err != nil {
		return err
	}

	// Initialize Gorp
	m.dbmap = &gorp.DbMap{
		Db:      db,
		Dialect: dialect,
	}
	// Create Tables for Messages
	m.CreateTables()
	err = m.dbmap.CreateTablesIfNotExists()
	if err != nil {
		return err
	}

	// Load the Server Keys
	loadedKey, err := identity.LoadKeyFromFile(m.KeyFile)
	if err != nil {
		loadedKey, err = identity.CreateIdentity()
		if err != nil {
			return err
		}
		if m.KeyFile != "" {
			err = loadedKey.SaveKeyToFile(m.KeyFile)
			if err != nil {
				return err
			}
		}
	}
	m.LogMessage("Loaded Address", loadedKey.Address.String())
	m.LogMessage("Loaded Encryption Key", hex.EncodeToString(crypto.RSAToBytes(loadedKey.Address.EncryptionKey)))

	loadedKey.SetLocation(m.Me)
	err = (&tracker.Router{
		URL:    m.TrackerURL,
		Origin: loadedKey,
	}).Register(loadedKey, m.Alias, nil)
	if err != nil {
		return err
	}

	m.Key = loadedKey

	handlers := []server.Handler{
		&dap.Handler{
			Key:      loadedKey,
			Delegate: m,
		},
	}

	// Create the AirDispatch Server
	adServer := server.Server{
		LocationName: m.Me,
		Key:          loadedKey,
		Delegate:     m,
		Handlers:     handlers,
		// Router:       mailserver.LookupRouter,
	}

	return adServer.StartServer(fmt.Sprintf("%d", port))
}

// SaveMessageDescription will serialize a message description and add it to the
// database.
func (m *Server) SaveMessageDescription(alert *message.EncryptedMessage) {
	err := m.SaveIncomingMessage(alert)
	if err != nil {
		m.HandleError(createError("Saving new alert to db.", err))
	}
}

// RetrieveMessageForUser will retrieve the message stored for the specified user
// at the specified name.
func (m *Server) RetrieveMessageForUser(name string, author *identity.Address, forAddr *identity.Address) *message.EncryptedMessage {
	msg, err := m.GetOutgoingMessageWithName(name, author.String(), forAddr.String())
	if err != nil {
		m.HandleError(createError("Getting message from DB.", err))
		return nil
	}

	data, err := msg.ToDispatch(forAddr.String())
	if err != nil {
		m.HandleError(createError("(RetrieveMessage) Marshalling message", err))
		return nil
	}

	return data
}

// RetrieveMessageListForUser will return all the public messages that a User
// has access to sent after `since`.
func (m *Server) RetrieveMessageListForUser(since uint64, author *identity.Address, forAddr *identity.Address) []*message.EncryptedMessage {
	results, err := m.GetOutgoingPublicMessagesFor(since, author.String(), forAddr.String())
	if err != nil {
		m.HandleError(createError("Getting messages from DB.", err))
		return nil
	}

	var out []*message.EncryptedMessage

	for _, v := range results {
		d, err := v.ToDispatch(forAddr.String())
		if err != nil {
			m.HandleError(createError("(RetrieveMessageList) Marshalling message", err))
			continue
		}
		out = append(out, d)
	}

	return out
}
