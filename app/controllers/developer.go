package controllers

import (
	"github.com/robfig/revel"
)

type Developer struct {
	*revel.Controller
}

func (d *Developer) Home() revel.Result {
	return d.Render()
}
