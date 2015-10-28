package router

import (
	"net/http"
	"strings"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/info"
)

// Router represents an HTTP multiplexer based on path prefixes.
type Router struct {
	globalPrefix string
	routes       map[string]Handler
}

// CreateRouter will create a new router that ignores the specified
// prefix.
func CreateRouter(globalPrefix string) *Router {
	return &Router{
		globalPrefix: globalPrefix,
		routes:       make(map[string]Handler),
	}
}

// AddRoute gives the router information about the routes it can
// serve.
func (r *Router) AddRoute(prefix string, handler interface{}) *Router {
	handlerInterface, handlerOk := handler.(Handler)
	getInterface, getOk := handler.(GetHandler)
	postInterface, postOk := handler.(PostHandler)

	if handlerOk {
		r.routes[prefix] = handlerInterface

		return r
	}

	if getOk || postOk {
		r.routes[prefix] = &selectionHandler{
			Get:  getInterface,
			Post: postInterface,
		}

		return r
	}

	panic("Cannot load a route that isn't a router.Handler, router.GetHandler, or router.PostHandler")
}

// Handle performs the actual routing.
func (r *Router) Handle(req *Request) framework.View {
	for route, handler := range r.routes {
		if strings.HasPrefix(req.URL.Path, r.globalPrefix+route) {
			return handler.Handle(req)
		}
	}

	return &framework.HTTPError{
		ErrorCode: 404,
		Message:   "Router is unable to found specified route.",
	}
}

func (r *Router) HandleRequest(req *http.Request, env *info.Environment) framework.View {
	return r.Handle(parseRequest(req, env))
}
