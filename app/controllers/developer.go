package controllers

import (
	"fmt"
	"github.com/airdispatch/dpl"
	"github.com/revel/revel"
	"html/template"
	"melange/mailserver"
	"net/http"
	"net/url"
	"runtime"
	"strings"
)

type Developer struct {
	*revel.Controller
}

func (d *Developer) Home() revel.Result {
	return d.Render()
}

func (d *Developer) Test(url string, action string, context string) (result revel.Result) {
	if url == "" && action == "" && context == "" {
		return d.Render()
	}

	defer func() {
		if r := recover(); r != nil {
			error := template.HTML(fmt.Sprintf("Recovered from Host Error %v (See Stack Trace)", r))

			var stackBytes []byte = make([]byte, 25000)
			n := runtime.Stack(stackBytes, true)
			stackBytes = stackBytes[:n]
			stack := string(stackBytes)

			result = d.Render(url, action, context, error, stack)
		}
	}()

	resp, err := http.Get(url)
	if err != nil {
		error := "Unable to load plugin from URL."
		return d.Render(url, action, context, error)
	}
	defer resp.Body.Close()

	o, err := dpl.ParseDPLStream(resp.Body)
	if err != nil {
		error := "Plugin is not valid."
		return d.Render(url, action, context, error)
	}

	plugin := o.CreateInstance(&DevelopHost{}, nil)

	display, err := plugin.RunAction(action)
	if err != nil {
		error := "Couldn't run plugin with context. Got error: " + err.Error()
		return d.Render(url, action, context, error)
	}

	result = d.Render(url, action, context, display)
	return
}

func (d *Developer) Address(a string) revel.Result {
	if a == "" {
		return d.Render()
	}
	mailserver.InitRouter()

	if strings.Contains(a, "@") {
		lookupType := "alias"
		addr, err := mailserver.LookupRouter.LookupAlias(a)
		return d.Render(a, addr, err, lookupType)
	} else {
		lookupType := "standard"
		addr, err := mailserver.LookupRouter.Lookup(a)
		return d.Render(a, addr, err, lookupType)
	}
	return d.Render(a)
}

type DevelopHost struct {
}

func (d *DevelopHost) GetMessages(plugin *dpl.PluginInstance, tag dpl.Tag, predicate *dpl.Predicate, limit int) ([]dpl.Message, error) {
	return nil, nil
}

func (d *DevelopHost) GetURLForAction(plugin *dpl.PluginInstance, action dpl.Action, message dpl.Message, user dpl.User) (*url.URL, error) {
	u, _ := url.Parse("http://google.com")
	return u, nil
}

func (d *DevelopHost) SendURL(plugin *dpl.PluginInstance) *url.URL {
	u, _ := url.Parse("/send")
	return u
}

func (d *DevelopHost) RunNotification(plugin *dpl.PluginInstance, n *dpl.Notification) {
	// What?
}

func (d *DevelopHost) Identify() string {
	return "Melange Developer v0.1"
}
