package framework

import (
	"io"
)

var Error403 = &HTTPError{403, "Authorization required."}
var Error404 = &HTTPError{404, "Resource not found."}
var Error500 = &HTTPError{500, "Internal service error."}

// Http Error
type HTTPError struct {
	ErrorCode int
	Message   string
}

func (h *HTTPError) Write(res io.Writer) {
	res.Write([]byte(h.Message))
}

func (h *HTTPError) ContentType() string {
	return "text/plain"
}

func (h *HTTPError) ContentLength() int {
	return len([]byte(h.Message))
}

func (h *HTTPError) Code() int {
	return h.ErrorCode
}

func (h *HTTPError) Headers() Headers { return nil }
