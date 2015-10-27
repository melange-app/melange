package tiles

import "getmelange.com/backend/api/router"

// Router holds all of the routes associated with the tile api module.
var Router = router.CreateRouter("tiles").
	AddRoute("/current", &CurrentTiles{}).
	AddRoute("/update", &UpdateTiles{})
