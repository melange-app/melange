package api

import (
	"net/http"

	// Import each of the API submodules in order to delegate
	// their routes and controllers.
	"getmelange.com/backend/api/identity"
	"getmelange.com/backend/api/marketplace"
	"getmelange.com/backend/api/messages"
	"getmelange.com/backend/api/people"
	"getmelange.com/backend/api/plugins"
	"getmelange.com/backend/api/router"
	"getmelange.com/backend/api/tiles"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/info"
	"getmelange.com/backend/realtime"
)

var optionsView = &framework.CORSView{
	View: &framework.RawView{
		Content: []byte(""),
		Type:    "text/plain",
	},
	Origin: "*",
}

// Specify the routes for the API module.
var routes = router.CreateRouter("").
	// api/identity
	AddRoute("identity", identity.Router).

	// api/marketplace
	AddRoute("market", marketplace.Router).

	// api/messages
	AddRoute("messages", messages.Router).
	AddRoute("profile", messages.ProfileRouter).

	// api/people
	AddRoute("people", people.Router).

	// api/plugins
	AddRoute("plugins", plugins.Router).
	AddRoute("update", plugins.UpdateRouter).

	// api/tiles
	AddRoute("tiles", tiles.Router)

// HandleRequest will redirect a request through the API Layer.
func HandleRequest(
	res http.ResponseWriter,
	req *http.Request,
	env *info.Environment,
) {
	// Don't process anything if the client just wants to see the
	// CORS headers.
	if req.Method == "OPTIONS" {
		return optionsView.Write(res)
	}

	// If the user is attempting to access /realtime, they want a
	// websockets connection.
	if req.URL.Path == "/realtime" {
		h := realtime.CreateRealtimeHandler(env)
		h.UpgradeConnection(res, req)
		return
	}

	routes.HandleRequest(req, env).Write(res)
}
