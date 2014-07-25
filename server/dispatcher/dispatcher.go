package dispatcher

import (
	"airdispat.ch/crypto"
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/server"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
	"melange/dap"
)

func CreateError(location string, error error) *server.ServerError {
	return &server.ServerError{
		Location: location,
		Error:    error,
	}
}

type Server struct {
	Me      string
	KeyFile string
	Key     string
	DBConn  string
	DBType  string
	dbmap   *gorp.DbMap
	server.BasicServer
}

func (s *Server) Run(port int) error {
	// Initialize Type of Database
	var dialect gorp.Dialect
	switch s.DBType {
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
	db, err := sql.Open(s.DBType, s.DBConn)
	if err != nil {
		return err
	}

	// Initialize Gorp
	s.dbmap = &gorp.DbMap{
		Db:      db,
		Dialect: dialect,
	}
	// Create Tables for Messages
	s.CreateTables()
	err = s.dbmap.CreateTablesIfNotExists()
	if err != nil {
		return err
	}

	// Load the Server Keys
	loadedKey, err := identity.LoadKeyFromFile(s.KeyFile)
	if err != nil {
		loadedKey, err = identity.CreateIdentity()
		if err != nil {
			return err
		}
		if s.KeyFile != "" {
			err = loadedKey.SaveKeyToFile(s.KeyFile)
			if err != nil {
				return err
			}
		}
	}
	s.LogMessage("Loaded Address", loadedKey.Address.String())
	s.LogMessage("Loaded Encryption Key", hex.EncodeToString(crypto.RSAToBytes(loadedKey.Address.EncryptionKey)))

	handlers := []server.Handler{
		&dap.Handler{
			Key:      loadedKey,
			Delegate: s,
		},
	}

	// Create the AirDispatch Server
	adServer := server.Server{
		LocationName: s.Me,
		Key:          loadedKey,
		Delegate:     s,
		Handlers:     handlers,
		// Router:       mailserver.LookupRouter,
	}

	return adServer.StartServer(fmt.Sprintf("%d", port))
}

func (m *Server) SaveMessageDescription(alert *message.EncryptedMessage) {
	err := m.SaveIncomingMessage(alert)
	if err != nil {
		m.HandleError(CreateError("Saving new alert to db.", err))
	}
}

func (m *Server) RetrieveMessageForUser(name string, author *identity.Address, forAddr *identity.Address) *message.EncryptedMessage {
	msg, err := m.GetOutgoingMessageWithName(name, author.String(), forAddr.String())
	if err != nil {
		m.HandleError(CreateError("Getting message from DB.", err))
		return nil
	}

	return msg.ToDispatch(forAddr.String())
}

func (m *Server) RetrieveMessageListForUser(since uint64, author *identity.Address, forAddr *identity.Address) []*message.EncryptedMessage {
	results, err := m.GetOutgoingPublicMessagesFor(since, author.String(), forAddr.String())
	if err != nil {
		m.HandleError(CreateError("Getting messages from DB.", err))
		return nil
	}

	out := make([]*message.EncryptedMessage, 0)

	for _, v := range results {
		d := v.ToDispatch(forAddr.String())
		out = append(out, d)
	}

	return out
}
