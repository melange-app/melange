package controllers

// This File Defines the TemplateTags that get passed into the each Template

import (
	"github.com/robfig/revel"
	"html/template"
)

func init() {
	// The `active` tag compares two values and outputs class='active' if they are equal
	revel.TemplateFuncs["active"] = func(title string, special string) template.HTMLAttr {
		if title == special {
			return template.HTMLAttr("class='active'")
		}
		return template.HTMLAttr("")
	}

	revel.OnAppStart(Init)
	revel.InterceptMethod((*GorpController).Begin, revel.BEFORE)
	revel.InterceptMethod((*GorpController).Commit, revel.AFTER)
	revel.InterceptMethod((*GorpController).Rollback, revel.FINALLY)
}
