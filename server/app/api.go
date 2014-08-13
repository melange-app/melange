package app

import (
	"fmt"
	"io"
	"melange/app/controllers"
	"melange/app/framework"
	"melange/app/models"
	"melange/app/packaging"
	"net/http"
	"os"
	"path/filepath"
)

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

// Handler is a struct that takes a request and returns a view.
type Handler interface {
	Handle(req *http.Request) framework.View
}

// HandleAPI will redirect a request through the API Layer.
func (r *Server) HandleAPI(res http.ResponseWriter, req *http.Request) {
	fmt.Println("API Request", req.Method, req.URL.Path)

	packager := &packaging.Packager{
		API: "http://www.getmelange.com/api",
	}

	tables, err := models.CreateTables(r.Settings)
	if err != nil {
		fmt.Println("Error creating tables", err)
		framework.WriteView(framework.Error500, res)
		return
	}

	pluginPath := filepath.Join(os.Getenv("MLGBASE"), "plugins")

	// Create Simple Handler Map
	handlers := map[string]Handler{
		//
		// PROVIDERS
		//

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

		//
		// PLUGINS
		//

		// GET  /plugins
		"/plugins": &PluginServer{
			Path: pluginPath,
		},

		//
		// TILES
		//

		"/tiles/current": &controllers.CurrentTiles{},
		"/tiles": &controllers.AllTiles{
			Path: pluginPath,
		},

		//
		// PROFILES
		//

		"/profile/current": &controllers.CurrentProfile{
			Store:  r.Settings,
			Tables: tables,
		},
		"/profile/update": &controllers.UpdateProfile{
			Store:  r.Settings,
			Tables: tables,
		},

		//
		// MESSAGES
		//

		// POST /messages
		"/messages": &controllers.Messages{
			Store:  r.Settings,
			Tables: tables,
		},
		// POST /messages/new
		"/messages/new": &controllers.NewMessage{
			Store:  r.Settings,
			Tables: tables,
		},
		// POST /messages/update
		"/messages/update": &controllers.UpdateMessage{
			Store:  r.Settings,
			Tables: tables,
		},

		//
		// DATA
		//

		// POST /data
		"/data": nil,
		// POST /data/set
		"/data/set": nil,

		//
		// CONTACTS
		//

		// GET  /contacts
		"/contacts": &controllers.ListContacts{
			Table: tables["contact"],
			Store: r.Settings,
		},
		// POST /contacts/new
		"/contacts/new": &controllers.NewContact{
			Tables: tables,
			Store:  r.Settings,
		},

		//
		// IDENTITY
		//

		// Get All Identities
		"/identity": &controllers.ListIdentity{
			Tables: tables,
			Store:  r.Settings,
		},
		// Create a new Identity
		"/identity/new": &controllers.SaveIdentity{
			Tables:   tables,
			Packager: packager,
			Store:    r.Settings,
		},
		// Current Identity Information
		"/identity/current": &controllers.CurrentIdentity{
			Tables: tables,
			Store:  r.Settings,
		},
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
	// Give them the 404, boss.
	framework.WriteView(
		&framework.HTTPError{
			ErrorCode: 404,
			Message:   "Couldn't get that page for you. Sorry.",
		}, res,
	)
}

// APIView will wrap a Handler and return CORS-compliant
// headers for the AppURL domain.
type APIView struct {
	AppURL string
	Handler
	framework.View
}

// Handle a request, don't call the inner handler if the request
// is of the OPTIONS method.
func (a *APIView) Handle(req *http.Request) framework.View {
	if req.Method != "OPTIONS" {
		a.View = a.Handler.Handle(req)
	}
	return a
}

// Write something if the method is not OPTIONS.
func (a *APIView) Write(r io.Writer) {
	if a.View != nil {
		a.View.Write(r)
	}
	return
}

// Code will return the status code if the method is not OPTIONS.
func (a *APIView) Code() int {
	if a.View != nil {
		return a.View.Code()
	}
	return 200
}

// ContentLength will return the content length if the method is not OPTIONS.
func (a *APIView) ContentLength() int {
	if a.View != nil {
		return a.View.ContentLength()
	}
	return 0
}

// ContentType will return the content type if the method is not OPTIONS.
func (a *APIView) ContentType() string {
	if a.View != nil {
		return a.View.ContentType()
	}
	return "text/plain"
}

// Headers add the CORS headers to the request:
//
// "Access-Control-Allow-Origin"
// "Access-Control-Allow-Headers"
func (a *APIView) Headers() framework.Headers {
	hdrs := make(framework.Headers)
	hdrs["Access-Control-Allow-Origin"] = a.AppURL
	hdrs["Access-Control-Allow-Headers"] = "Content-Type"
	return hdrs
}

// POSTHandler contains an inner handler that it utilizes only if
// the request is the POST method.
type POSTHandler struct {
	Handler
}

// Handle will redirect to the inner handler only if the request is
// the POST method.
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

// PluginServer returns the list of plugins from the MLGBASE/plugins directory.
type PluginServer struct {
	Path string
}

// Handle the PluginServer.
func (s *PluginServer) Handle(req *http.Request) framework.View {
	return framework.LoadPlugins(s.Path)
}

// ServerLists will return the list of Trackers or Servers from
// getmelange.com.
type ServerLists struct {
	URL      string
	Packager *packaging.Packager
}

// Handle will decodeProviders from getmelange.com then return them in JSON.
func (s *ServerLists) Handle(req *http.Request) framework.View {
	packages, err := s.Packager.DecodeProviders(s.Packager.API + s.URL)
	if err != nil {
		fmt.Println(err)
		return framework.Error500
	}
	return &framework.JSONView{
		Content: packages,
	}
}
