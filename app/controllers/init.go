package controllers

// This File Defines the TemplateTags that get passed into the each Template

import (
	"github.com/robfig/revel"
	"html/template"
	"time"
)

func init() {
	// The `active` tag compares two values and outputs class='active' if they are equal
	revel.TemplateFuncs["active"] = func(title string, special string) template.HTMLAttr {
		if title == special {
			return template.HTMLAttr("class='active'")
		}
		return template.HTMLAttr("")
	}

	revel.TemplateFuncs["since"] = func(t time.Time) string {
		/*n := time.Now().Truncate(time.Second)
		if SameDay(n, t) {
			return fmt.Sprintf("%v hours ago", n.Truncate(time.Hour).Sub(t.Truncate(time.Hour)).String())
		} else if SameDay(n.Add(-24*time.Hour), t) {
			return "Yesterday"
		} else {
			return fmt.Sprintf("%d days ago", n.Sub(t)/(24*time.Hour))
		}*/
		return t.Format("Jan 2 at 3:04 PM")
	}

	revel.OnAppStart(Init)

	revel.InterceptMethod((*GorpController).Begin, revel.BEFORE)
	revel.InterceptMethod((*GorpController).Commit, revel.AFTER)
	revel.InterceptMethod((*GorpController).Rollback, revel.FINALLY)

	revel.InterceptMethod((*Dispatch).Init, revel.BEFORE)
}

func SameDay(t1 time.Time, t2 time.Time) bool {
	return false
}
