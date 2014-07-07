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
	"net"
	"os"
	"time"
)

func getServerLocation(port int) string {
	s, _ := os.Hostname()
	ips, _ := net.LookupHost(s)
	return fmt.Sprintf("%s:%d", ips[0], port)
}

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

	// Initialize the Router (Uneeded Here)
	// mailserver.InitRouter()

	// Create the AirDispatch Server
	adServer := server.Server{
		LocationName: s.Me,
		Key:          loadedKey,
		Delegate:     s,
		// Router:       mailserver.LookupRouter,
	}

	return theServer.StartServer(*port)

	// This probably needs to be removed.
	// for {
	// 	fmt.Print("ad> ")
	// 	var prompt string
	// 	fmt.Scan(&prompt)
	// 	if prompt == "quit" {
	// 		return
	// 	}
	// }
}

func (m *Server) SaveMessageDescription(alert *server.MessageDescription) {
	a := models.CreateAlertFromDescription(alert)
	err := m.Map.Insert(a)
	if err != nil {
		m.HandleError(CreateError("Saving new alert to db.", err))
	}
}

func (m *Server) RetrieveMessageForUser(name string, author *identity.Address, forAddr *identity.Address) *message.Mail {
	if name == "profile" {
		var id []*models.Identity
		_, err := m.Map.Select(&id, "select * from dispatch_identity where fingerprint = $1", author.String())
		if err != nil {
			m.HandleError(CreateError("Loading Profile Identity", err))
			return nil
		}
		if len(id) != 1 {
			// Couldn't find user
			m.HandleError(CreateError("Finding Profile User", errors.New(fmt.Sprintf("Found number of ids %v", len(id)))))
			return nil
		}
		u, err := m.Map.Get(&models.User{}, id[0].UserId)
		if err != nil {
			m.HandleError(CreateError("Getting Profile User", err))
			return nil
		}
		user, ok := u.(*models.User)
		if !ok {
			return nil
		}

		did, err := id[0].ToDispatch()
		if err != nil {
			m.HandleError(CreateError("Changing ID To Dispatch Profile", err))
			return nil
		}

		m := message.CreateMail(did.Address, forAddr, time.Now())
		m.Components.AddComponent(message.CreateComponent("airdispat.ch/profile/name", []byte(user.Name)))
		m.Components.AddComponent(message.CreateComponent("airdispat.ch/profile/avatar", []byte(user.GetAvatar())))

		return m
	}

	var results []*models.Message
	_, err := m.Map.Select(&results, "select * from dispatch_messages where name = $1 and \"from\" = $2", name, author.String())
	if err != nil {
		m.HandleError(CreateError("Finding Messages for User", err))
		return nil
	}
	if len(results) != 1 {
		m.HandleError(CreateError("Finding Messages for User", errors.New("Found wrong number of messages.")))
		return nil
	}

	out, err := results[0].ToDispatch(m.Map, forAddr.String())
	if err != nil {
		m.HandleError(CreateError("Creating Dispatch Message", err))
		return nil
	}
	return out
}

func (m *Server) RetrieveMessageListForUser(since uint64, author *identity.Address, forAddr *identity.Address) *server.MessageList {
	var results []*models.Message
	_, err := m.Map.Select(&results, "select * from dispatch_messages where \"from\" = $1 and timestamp > $2 and \"to\" = ''", author.String(), since)
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
		desc.Location = *me
		out.AddMessageDescription(desc)
	}

	return out
}
