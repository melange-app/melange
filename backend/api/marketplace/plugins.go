package marketplace

import (
	"fmt"
	"net/http"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/packaging"
)

// ServerLists will return the list of Trackers or Servers from
// getmelange.com.
type PluginStoreController struct {
	Packager *packaging.Packager
}

// Handle will decodeProviders from getmelange.com then return them in JSON.
func (s *PluginStoreController) Handle(req *http.Request) framework.View {
	packages, err := s.Packager.DecodeStore()
	if err != nil {
		fmt.Println("Getting store", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: packages,
	}
}
