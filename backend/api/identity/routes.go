package identity

import "getmelange.com/backend/api/router"

// Router represents the multiplexer for requests that come into the
// identity module of the API.
var Router = router.CreateRouter("/identity").
	AddRoute("/new", &SaveIdentity{}).
	AddRoute("/current", &CurrentIdentity{}).
	AddRoute("/remove", &RemoveIdentity{}).
	AddRoute("", &ListIdentity{})
