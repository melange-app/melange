package app

import (
	"airdispat.ch/identity"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/huntaub/go-db"
	"io"
	"melange/app/framework"
	"melange/app/models"
	"melange/dap"
	"melange/router"
	"net/http"
	"os"
	"path/filepath"
)

// Create a CORS View to be accessed by the Application URL
type APIView struct {
	AppURL string
	Handler
	framework.View
}

func (a *APIView) Handle(req *http.Request) framework.View {
	if req.Method != "OPTIONS" {
		a.View = a.Handler.Handle(req)
	}
	return a
}

func (a *APIView) Write(r io.Writer) {
	if a.View != nil {
		a.View.Write(r)
	}
	return
}

func (a *APIView) Code() int {
	if a.View != nil {
		return a.View.Code()
	}
	return 200
}

func (a *APIView) ContentLength() int {
	if a.View != nil {
		return a.View.ContentLength()
	}
	return 0
}

func (a *APIView) ContentType() string {
	if a.View != nil {
		return a.View.ContentType()
	}
	return "text/plain"
}

func (a *APIView) Headers() framework.Headers {
	hdrs := make(framework.Headers)
	hdrs["Access-Control-Allow-Origin"] = a.AppURL
	hdrs["Access-Control-Allow-Headers"] = "Content-Type"
	return hdrs
}

func DecodeJSONBody(req *http.Request, object interface{}) error {
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()
	return decoder.Decode(object)
}

// API Functions
//
// POST /register
// POST /unregister
//
// POST /messages
// POST /messages/new
// POST /messages/update
//
// POST /data
// POST /data/set

type Handler interface {
	Handle(req *http.Request) framework.View
}

const MelangeSite = "http://localhost:3000"

func (r *Server) HandleApi(res http.ResponseWriter, req *http.Request) {
	packager := &Packager{
		API: "http://www.getmelange.com/api",
	}

	tables, err := models.CreateTables(r.Settings)
	if err != nil {
		fmt.Println("Error creating tables", err)
		framework.WriteView(framework.Error500, res)
		return
	}

	// Create Simple Handler Map
	handlers := map[string]Handler{
		// GET  /servers
		"/servers": &ServerLists{
			URL:      "/servers",
			Packager: packager,
		},
		// GET  /trackers
		"/trackers": &ServerLists{
			URL:      "/trackers",
			Packager: packager,
		},

		// GET  /plugins
		"/plugins": &PluginServer{},

		// POST /messages
		"/messages": nil,
		// POST /messages/new
		"/messages/new": nil,
		// POST /messages/update
		"/messages/update": nil,

		// POST /data
		"/data": nil,
		// POST /data/set
		"/data/set": nil,

		"/identity": nil,
		"/identity/new": &Identity{
			Tables:   tables,
			Packager: packager,
			Store:    r.Settings,
		},

		// POST /register
		"/identity/reigster": &Register{true},
		// POST /unregister
		"/identity/unregister": &Register{false},
	}

	// Run through API Handlers
	for route, handler := range handlers {
		if req.URL.Path == route {
			view := (&APIView{
				AppURL:  r.AppURL(),
				Handler: handler,
			}).Handle(req)
			framework.WriteView(
				view, res,
			)
			return
		}
	}
}

type POSTHandler struct {
	Handler
}

func (p *POSTHandler) Handle(req *http.Request) framework.View {
	// If method is not POST, disallow request.
	if req.Method != "POST" {
		return &framework.HTTPError{
			ErrorCode: 405,
			Message:   "Method not allowed.",
		}
	}
	return p.Handler.Handle(req)
}

type PluginServer struct{}

func (s *PluginServer) Handle(req *http.Request) framework.View {
	return framework.LoadPlugins(filepath.Join(os.Getenv("MLGBASE"), "plugins"))
}

type ServerLists struct {
	URL      string
	Packager *Packager
}

func (s *ServerLists) Handle(req *http.Request) framework.View {
	packages, err := s.Packager.decodeProviders(s.Packager.API + s.URL)
	if err != nil {
		fmt.Println(err)
		return framework.Error500
	}
	return &framework.JSONView{
		Content: packages,
	}
}

type Register struct {
	Starting bool
}

func (r *Register) Handle(req *http.Request) framework.View {
	unmarshalled := make(map[string]string)
	err := DecodeJSONBody(req, &unmarshalled)
	if err != nil {
		return framework.Error500
	}

	uploadable := make(map[string][]byte)
	for key, value := range unmarshalled {
		uploadable[key] = []byte(value)
	}

	return nil
}

type Profile struct {
	// Profile Information
	FirstName string `json:"first"`
	LastName  string `json:"last"`
	About     string `json:"about"`
	Password  string `json:"password"`
	// AD Information
	Server  string `json:"server"`
	Tracker string `json:"tracker"`
	Alias   string `json:"alias"`
	// Profile Nickname
	Nickname string `json:"nickname"`
}

// identity
// identity:_addr_:key
// identity:_addr_:location
type Identity struct {
	Tables   map[string]db.Table
	Store    *models.Store
	Packager *Packager
}

func (i *Identity) Handle(req *http.Request) framework.View {
	// Decode Body
	profileRequest := &Profile{}
	err := DecodeJSONBody(req, &profileRequest)

	if err != nil && err != io.EOF {
		fmt.Println("Error occured while decoding body:", err)
		return framework.Error500
	}

	// Create Identity
	id, err := identity.CreateIdentity()
	if err != nil {
		fmt.Println("Error occured creating an identity:", err)
		return framework.Error500
	}

	//
	// Server Registration
	//

	// Extract Keys
	server, err := i.Packager.ServerFromId(profileRequest.Server)
	if err != nil {
		fmt.Println("Error occured getting server:", err)
		return &framework.HTTPError{
			ErrorCode: 500,
			Message:   "Couldn't get server for id.",
		}
	}

	// Run Registration
	client := &dap.Client{
		Key:    id,
		Server: server.Key,
	}
	err = client.Register(map[string][]byte{
		"name": []byte(profileRequest.FirstName + " " + profileRequest.LastName),
	})
	if err != nil {
		fmt.Println("Error occurred registering on Server", err)
		return framework.Error500
	}

	//
	// Tracker Registration
	//

	tracker, err := i.Packager.TrackerFromId(profileRequest.Tracker)
	if err != nil {
		fmt.Println("Error occured getting tracker:", err)
		return &framework.HTTPError{
			ErrorCode: 500,
			Message:   "Couldn't get tracker for id.",
		}
	}

	err = (&router.Router{
		Origin: id,
		TrackerList: []string{
			tracker.URL,
		},
	}).Register(id, profileRequest.Alias)

	if err != nil {
		fmt.Println("Error occurred registering on Tracker", err)
		return framework.Error500
	}

	//
	// Database Registration
	//

	buffer := &bytes.Buffer{}
	_, err = id.GobEncodeKey(buffer)
	if err != nil {
		fmt.Println("Error occurred encoding Id", err)
		return framework.Error500
	}

	modelId := &models.Identity{
		Nickname:    profileRequest.Nickname,
		Fingerprint: id.Address.String(),
		Data:        buffer.Bytes(),
		Protected:   false,
		Server:      server.URL,
	}

	_, err = i.Tables["identity"].Insert(modelId).Exec(i.Store)
	if err != nil {
		fmt.Println("Error saving Identity", err)
		return framework.Error500
	}

	modelAlias := &models.Alias{
		Identity: db.ForeignKey(modelId),
		Location: tracker.URL,
		Username: profileRequest.Alias,
	}

	_, err = i.Tables["alias"].Insert(modelAlias).Exec(i.Store)
	if err != nil {
		fmt.Println("Error saving Alias", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: map[string]interface{}{
			"error": false,
		},
	}
}
