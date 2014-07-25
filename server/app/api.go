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
	fmt.Println("Entire the API", req.Method, req.URL.Path)

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

		"/identity": &ListIdentity{
			Tables: tables,
			Store:  r.Settings,
		},
		"/identity/new": &SaveIdentity{
			Tables:   tables,
			Packager: packager,
			Store:    r.Settings,
		},
		"/identity/current": &CurrentIdentity{
			Tables: tables,
			Store:  r.Settings,
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

func CurrentIdentityOrError(store *models.Store, table db.Table) (*models.Identity, *framework.HTTPError) {
	data, err := store.GetDefault("current_identity", "")
	if err != nil {
		fmt.Println("Error getting current identity.", err)
		return nil, framework.Error500
	}

	if data == "" {
		return nil, &framework.HTTPError{
			ErrorCode: 422,
			Message:   "Cannot fufill current identity request before creating an identity.",
		}
	}

	result := &models.Identity{}
	err = table.Get().Where("fingerprint", data).One(store, result)
	if err != nil {
		fmt.Println("Error getting identity:", err)
		return nil, framework.Error500
	}

	return result, nil
}

type CurrentIdentity struct {
	Tables map[string]db.Table
	Store  *models.Store
}

func (i *CurrentIdentity) Handle(req *http.Request) framework.View {
	if req.Method == "POST" {
		request := make(map[string]interface{})
		err := DecodeJSONBody(req, &request)
		if err != nil {
			fmt.Println("Cannot decode body", err)
			return framework.Error500
		}

		fingerprint, ok := request["fingerprint"]
		if !ok {
			return &framework.HTTPError{
				ErrorCode: 400,
				Message:   "Fingerprint is required.",
			}
		}

		err = i.Store.Set("current_identity", fingerprint.(string))
		if err != nil {
			fmt.Println("Error storing current_identity.", err)
			return framework.Error500
		}

		return &framework.JSONView{
			Content: map[string]interface{}{
				"error": false,
			},
		}

	} else if req.Method == "GET" {
		result, err := CurrentIdentityOrError(i.Store, i.Tables["identity"])
		if err != nil {
			return err
		}

		return &framework.JSONView{
			Content: result,
		}
	}
	return &framework.HTTPError{
		ErrorCode: 405,
		Message:   "Method not allowed.",
	}
}

type ListIdentity struct {
	Tables map[string]db.Table
	Store  *models.Store
}

func (i *ListIdentity) Handle(req *http.Request) framework.View {
	var results []*models.Identity
	err := i.Tables["identity"].Get().All(i.Store, &results)
	if err != nil {
		fmt.Println("Error getting Identities:", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: results,
	}
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

type SaveIdentity struct {
	Tables   map[string]db.Table
	Store    *models.Store
	Packager *Packager
}

func (i *SaveIdentity) Handle(req *http.Request) framework.View {
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

	// Save as the current identity.
	err = i.Store.Set("current_identity", id.Address.String())
	if err != nil {
		fmt.Println("Error storing current_identity.", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: map[string]interface{}{
			"error": false,
		},
	}
}
