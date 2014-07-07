package dispatcher

import (
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/server"
	"database/sql"
	"errors"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
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

	// Load the Server Keys
	loadedKey, err := identity.LoadKeyFromFile(s.KeyFile)
	if err != nil {
		loadedKey, err = identity.CreateIdentity()
		if err != nil {
			return err
		}
		if s.Key != "" {
			err = loadedKey.SaveKeyToFile(s.KeyFile)
			if err != nil {
				return err
			}
		}
	}
	s.LogMessage("Loaded Address", loadedKey.Address.String())

	// Create the AirDispatch Server
	adServer := server.Server{
		LocationName: s.Me,
		Key:          loadedKey,
		Delegate:     s,
		// Router:       mailserver.LookupRouter,
	}

	return adServer.StartServer(fmt.Sprintf("%d", port))
}

func (m *Server) SaveMessageDescription(alert *server.MessageDescription) {
	a := CreateAlertFromDescription(alert)
	err := m.dbmap.Insert(a)
	if err != nil {
		m.HandleError(CreateError("Saving new alert to db.", err))
	}
}

func (m *Server) RetrieveMessageForUser(name string, author *identity.Address, forAddr *identity.Address) *message.Mail {
	var results []*OutgoingMessage
	_, err := m.dbmap.Select(&results, "select * from dispatch_messages where name = $1 and \"from\" = $2", name, author.String())
	if err != nil {
		m.HandleError(CreateError("Finding Messages for User", err))
		return nil
	}
	if len(results) != 1 {
		m.HandleError(CreateError("Finding Messages for User", errors.New("Found wrong number of messages.")))
		return nil
	}

	out, err := results[0].ToDispatch(m.dbmap, forAddr.String())
	if err != nil {
		m.HandleError(CreateError("Creating Dispatch Message", err))
		return nil
	}
	return out
}

func (m *Server) RetrieveMessageListForUser(since uint64, author *identity.Address, forAddr *identity.Address) *server.MessageList {
	var results []*OutgoingMessage
	_, err := m.dbmap.Select(&results, "select * from dispatch_messages where \"from\" = $1 and timestamp > $2 and \"to\" = ''", author.String(), since)
	if err != nil {
		m.HandleError(CreateError("Loading messages from DB", err))
		return nil
	}

	out := server.CreateMessageList(author, forAddr)

	for _, v := range results {
		desc, err := v.ToDescription(forAddr.String())
		if err != nil {
			m.HandleError(CreateError("Loading description.", err))
			return nil
		}
		desc.Location = m.Me
		out.AddMessageDescription(desc)
	}

	return out
}
