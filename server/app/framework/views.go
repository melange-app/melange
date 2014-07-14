package framework

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

func WriteView(view View, res http.ResponseWriter) {
	// Set Headers
	res.Header().Add("Content-Type", view.ContentType())
	contentLength := strconv.Itoa(view.ContentLength())
	res.Header().Add("Content-Length", contentLength)

	for key, value := range view.Headers() {
		res.Header().Add(key, value)
	}

	res.WriteHeader(view.Code())

	view.Write(res)
}

type Headers map[string]string

type View interface {
	Write(io.Writer)
	ContentType() string
	ContentLength() int
	Code() int
	Headers() Headers
}

type RawView struct {
	Content []byte
	Type    string
}

func (j *RawView) Write(w io.Writer) {
	w.Write(j.Content)
}

func (j *RawView) Code() int           { return 200 }
func (j *RawView) ContentLength() int  { return len(j.Content) }
func (j *RawView) ContentType() string { return j.Type }
func (j *RawView) Headers() Headers    { return nil }

type JSONView struct {
	Content interface{}
	cache   []byte
}

func (j *JSONView) getCache() []byte {
	if j.cache != nil {
		return j.cache
	}
	bytes, err := json.Marshal(j.Content)
	if err != nil {
		panic("JSON Encoding " + err.Error())
	}
	j.cache = bytes
	return j.cache
}

func (j *JSONView) Write(w io.Writer) {
	w.Write(j.getCache())
}

func (j *JSONView) Code() int           { return 200 }
func (j *JSONView) ContentLength() int  { return len(j.getCache()) }
func (j *JSONView) ContentType() string { return "application/json" }
func (j *JSONView) Headers() Headers    { return nil }
