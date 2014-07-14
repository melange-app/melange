package app

import (
	"encoding/json"
	"melange/app/framework"
	// "melange/dap"
	"airdispat.ch/identity"
	"fmt"
	"net/http"
  "os"
  "path/filepath"
)

// Create a CORS View to be accessed by the Application URL
type APIView struct {
	AppURL string
	framework.View
}

func (a *APIView) Headers() framework.Headers {
	hdrs := a.View.Headers()
	if hdrs == nil {
		hdrs = make(framework.Headers)
	}

	hdrs["Access-Control-Allow-Origin"] = a.AppURL
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

func (r *Server) HandleApi(res http.ResponseWriter, req *http.Request) {

	// Create Simple Handler Map
	handlers := map[string]Handler{
		// GET  /servers
		"/servers": &ServerLists{0},
		// GET  /trackers
		"/trackers": &ServerLists{1},

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

		"/identity":     nil,
		"/identity/new": &Identity{r.Settings},

		// POST /register
		"/identity/reigster": &Register{true},
		// POST /unregister
		"/identity/unregister": &Register{false},
	}

	// Run through API Handlers
	for route, handler := range handlers {
		if req.URL.Path == route {
			view := handler.Handle(req)
			framework.WriteView(
				&APIView{r.AppURL(), view}, res,
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
	Type int
}

func (s *ServerLists) Handle(req *http.Request) framework.View {
	var servers []*Provider
	if s.Type == 0 {
		servers = GetServers()
	} else if s.Type == 1 {
		servers = GetTrackers()
	}

	return &framework.JSONView{
		Content: servers,
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
	FirstName string
	LastName  string
	About     string
	Password  string
	// AD Information
	Server  string
	Tracker string
	Alias   string
}

// identity
// identity:_addr_:key
// identity:_addr_:location
type Identity struct {
	Settings *Store
}

func (i *Identity) Handle(req *http.Request) framework.View {
	profileRequest := &Profile{}
	err := DecodeJSONBody(req, &profileRequest)
	if err != nil {
		return framework.Error500
	}

	// Create Identity
	_, err = identity.CreateIdentity()
	if err != nil {
		fmt.Println("Error occured creating an identity: ", err)
		return framework.Error500
	}

	return nil
}
