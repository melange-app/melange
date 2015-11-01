package marketplace

import "getmelange.com/backend/api/router"

// Router holds all of the routes associated with the marketplace api
// module.
var Router = router.CreateRouter("/market").
	AddRoute("/servers", &ServerLists{})
