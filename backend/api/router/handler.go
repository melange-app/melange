package router

import "getmelange.com/backend/framework"

// GetHandler is a type of Handler that only accepts GET requests.
type GetHandler interface {
	Get(
		req *Request,
	) framework.View
}

// PostHandler is a type of Handler that only accepts POST requests.
type PostHandler interface {
	Post(
		req *Request,
	) framework.View
}

// Handler is a type of object that returns a framework.View after it
// completes the rendering.
type Handler interface {
	Handle(
		req *Request,
	) framework.View
}
