package marketplace

import (
	"fmt"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"
	"getmelange.com/backend/packaging"
)

// ServerLists will return the list of Trackers or Servers from
// getmelange.com.
type ServerLists struct {
	URL      string
	Packager *packaging.Packager
}

// Handle will decodeProviders from getmelange.com then return them in JSON.
func (s *ServerLists) Get(req *router.Request) framework.View {
	extra := s.URL
	if s.Packager.Debug {
		extra += "?debug=true"
	}

	packages, err := s.Packager.GetServers()
	if err != nil {
		fmt.Println(err)
		return framework.Error500
	}
	return &framework.JSONView{
		Content: packages,
	}
}
