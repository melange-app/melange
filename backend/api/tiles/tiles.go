package tiles

import (
	"fmt"
	"net/http"
	"strings"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/models"
)

type CurrentTiles struct {
	Store *models.Store
}

func (c *CurrentTiles) Handle(req *http.Request) framework.View {
	data, err := c.Store.GetDefault("tiles_current", "")
	if err != nil {
		fmt.Println("Error getting from the store", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: strings.Split(data, ","),
	}
}

type UpdateTiles struct {
	Store *models.Store
}

func (u *UpdateTiles) Handle(req *http.Request) framework.View {
	// Decode Body
	tiles := make([]string, 0)
	err := DecodeJSONBody(req, &tiles)

	if err != nil {
		fmt.Println("Error occured while decoding body:", err)
		return framework.Error500
	}

	err = u.Store.Set("tiles_current", strings.Join(tiles, ","))
	if err != nil {
		fmt.Println("Error setting current tiles", err)
		return framework.Error500
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "OK",
	}
}
