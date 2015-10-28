package router

import (
	"encoding/json"
	"net/http"

	"getmelange.com/backend/info"
)

// Request is an HTTP request with the important fields already parsed
// out and ready for use.
type Request struct {
	Environment *info.Environment

	*http.Request
}

// JSON will parse the request body and return the deserialized JSON
// object.
func (r *Request) JSON(obj interface{}) error {
	decoder := json.NewDecoder(r.Request.Body)
	defer r.Request.Body.Close()

	return decoder.Decode(obj)
}

func parseRequest(req *http.Request, env *info.Environment) *Request {
	return &Request{
		Request:     req,
		Environment: env,
	}
}
