package tiles

import (
	"fmt"
	"strings"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"
)

// CurrentTiles will retrieve the current set of tiles to be displayed
// on the home page.
type CurrentTiles struct{}

// Get will execute the retrieval.
func (c *CurrentTiles) Get(req *router.Request) framework.View {
	data, err := req.Environment.Store.GetDefault("tiles_current", "")
	if err != nil {
		fmt.Println("Error getting from the store", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: strings.Split(data, ","),
	}
}

// UpdateTiles will update the current set of tiles displayed to the
// user.
type UpdateTiles struct{}

// Post will execute the update.
func (u *UpdateTiles) Post(req *router.Request) framework.View {
	tiles := make([]string, 0)
	err := req.JSON(&tiles)

	if err != nil {
		fmt.Println("Error occured while decoding body:", err)
		return framework.Error500
	}

	err = req.Environment.Store.Set("tiles_current",
		strings.Join(tiles, ","))
	if err != nil {
		fmt.Println("Error setting current tiles", err)
		return framework.Error500
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "OK",
	}
}
