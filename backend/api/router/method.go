package router

import (
	"net/http"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/info"
)

type selectionHandler struct {
	Get  GetHandler
	Post PostHandler
}

func (s *selectionHandler) Handle(req *http.Request, env *info.Environment) framework.View {
	if req.Method == "GET" && s.Get != nil {
		return s.Get.Get(req, env)
	} else if req.Method == "POST" && s.Post != nil {
		return s.Post.Post(req, env)
	}

	return &framework.HTTPError{
		ErrorCode: 405,
		Message:   "Method not allowed.",
	}
}
