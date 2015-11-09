package router

import "getmelange.com/backend/framework"

type selectionHandler struct {
	Get  GetHandler
	Post PostHandler
}

func (s *selectionHandler) Handle(req *Request) framework.View {
	if req.Method == "GET" && s.Get != nil {
		return s.Get.Get(req)
	} else if req.Method == "POST" && s.Post != nil {
		return s.Post.Post(req)
	}

	return &framework.HTTPError{
		ErrorCode: 405,
		Message:   "Method not allowed.",
	}
}
