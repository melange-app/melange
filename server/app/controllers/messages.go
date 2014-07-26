package controllers

import (
	"melange/app/framework"
	"melange/app/models"
	"net/http"

	gdb "github.com/huntaub/go-db"
)

type Messages struct {
	Store  *models.Store
	Tables map[string]gdb.Table
}

func (m *Messages) Handle(req *http.Request) framework.View {
	return &framework.HTTPError{
		ErrorCode: 504,
		Message:   "Not implemented.",
	}
}

type UpdateMessage struct {
	Store  *models.Store
	Tables map[string]gdb.Table
}

func (m *UpdateMessage) Handle(req *http.Request) framework.View {
	return &framework.HTTPError{
		ErrorCode: 504,
		Message:   "Not implemented.",
	}
}

type NewMessage struct {
	Store  *models.Store
	Tables map[string]gdb.Table
}

func (m *NewMessage) Handle(req *http.Request) framework.View {
	return &framework.HTTPError{
		ErrorCode: 504,
		Message:   "Not implemented.",
	}
}
