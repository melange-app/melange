package people

import "getmelange.com/backend/api/router"

var listRouter = router.CreateRouter("/people/list").
	AddRoute("/add", &AddList{}).
	AddRoute("/remove", &RemoveList{}).
	AddRoute("", &GetLists{})

// Router holds all of the routes associated with the people api
// module.
var Router = router.CreateRouter("/people").
	AddRoute("/update", &UpdateContact{}).
	AddRoute("/remove", &RemoveContact{}).
	AddRoute("/add", &AddContact{}).
	AddRoute("/list", listRouter).
	AddRoute("", &ListContacts{})
