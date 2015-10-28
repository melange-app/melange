package messages

import "getmelange.com/backend/api/router"

// ProfileRouter holds all of the routes associated with the profile
// portion of the messages api module.
var ProfileRouter = router.CreateRouter("profile").
	AddRoute("/current", &CurrentProfile{}).
	AddRoute("/update", &UpdateProfile{})

// Router holds all of the possible routes for the messages api
// module.
var Router = router.CreateRouter("messages").
	AddRoute("/get", &GetMessage{}).
	AddRoute("/at", &GetAllMessagesAt{}).
	AddRoute("/new", &NewMessage{}).
	AddRoute("/update", &UpdateMessage{}).
	AddRoute("/", &Messages{})
