package controllers

import (
	"airdispat.ch/routing"
	"fmt"
	"github.com/airdispatch/dpl"
	"github.com/coopernurse/gorp"
	"github.com/robfig/revel"
	"melange/app/models"
	"melange/app/routes"
	"melange/mailserver"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Loader struct {
	Dispatch
}

type PluginHost struct {
	User *models.User
	Txn  *gorp.Transaction
	R    routing.Router
}

func (p *PluginHost) GetMessages(plugin *dpl.PluginInstance, tag dpl.Tag, predicate *dpl.Predicate, limit int) ([]dpl.Message, error) {
	// Public Messages
	fromSub, err := models.UserIdentities(p.User, p.Txn)
	if err != nil {
		return nil, err
	}

	return mailserver.Messages(p.R, p.Txn, fromSub[0], p.User, true, true, true, time.Now().Add(-7*24*time.Hour).Unix())
}

func (p *PluginHost) SendURL(plugin *dpl.PluginInstance) *url.URL {
	rawURL := fmt.Sprintf("/app/%s/send", strings.ToLower(plugin.Name))
	u, _ := url.Parse(rawURL)
	return u
}

func (p *PluginHost) GetURLForAction(
	plugin *dpl.PluginInstance, // The Plugin Calling for the URL
	action dpl.Action, // The Action
	message dpl.Message, // MessageContext
	user dpl.User) (*url.URL, error) { // User Context

	rawURL := fmt.Sprintf("/app/%s/%s", strings.ToLower(plugin.Name), action.Name)
	if message != nil {
		rawURL = fmt.Sprintf("%s/m/%d", rawURL, message.Created().Unix)
	}
	if user != nil {
		rawURL = fmt.Sprintf("%s/u/%s", rawURL, user.Name)
	}

	url, _ := url.Parse(rawURL)

	return url, nil
}

func (p *PluginHost) RunNotification(plugin *dpl.PluginInstance, n *dpl.Notification) {
	// Log Notification Reception
}

func (p *PluginHost) Identify() string {
	return "Melange v0.1"
}

func (d Loader) LoadAppDefault(app string) revel.Result {
	return d.LoadApp(app, "", nil, nil)
}

func (d Loader) LoadAppAction(app string, action string) revel.Result {
	return d.LoadApp(app, action, nil, nil)
}

func (d Loader) LoadAppMessage(app string, action string, message int) revel.Result {
	d.SetAction("Loader", "LoadApp")
	return d.LoadApp(app, action, nil, nil)
}

func (d Loader) LoadAppUser(app string, action string, username string) revel.Result {
	d.SetAction("Loader", "LoadApp")
	return d.LoadApp(app, action, nil, nil)
}

// We should load the Application just once from the XML, then refresh it at a regular interval
func (d Loader) LoadApp(app string, action string, message dpl.Message, user dpl.User) revel.Result {
	var apps []*models.UserApp
	_, err := d.Txn.Select(&apps, "select * from dispatch_app where userid = $1 and UPPER(name) = UPPER($2)", GetUserId(d.Session), app)
	if err != nil {
		panic(err)
	}
	if len(apps) != 1 {
		panic(len(apps))
	}

	resp, err := http.Get(apps[0].AppURL)
	if err != nil {
		d.RenderArgs["error"] = "We were unable to load this application from the supplied URL. Maybe it has moved or the URL was incorrect?"
		d.RenderArgs["app"] = apps[0]

		return d.RenderTemplate("loader/error.html")
	}
	defer resp.Body.Close()

	// The Plugin is O
	o, err := dpl.ParseDPLStream(resp.Body)
	if err != nil {
		d.RenderArgs["error"] = "This plugin is no longer valid. You should contact the plugin developer."
		d.RenderArgs["app"] = apps[0]

		return d.RenderTemplate("loader/error.html")
	}

	// Get Current User
	u, err := d.Txn.Get(&models.User{}, GetUserId(d.Session))
	if err != nil {
		panic(err)
	}

	// Create a Singleton that Hosts all the Plugins
	GlobalHost := &PluginHost{
		User: u.(*models.User),
		Txn:  d.Txn,
		R:    mailserver.LookupRouter,
	}

	plugin := o.CreateInstance(GlobalHost, nil)

	data, err := plugin.RunActionWithContext(action, message, user)
	if err != nil {
		panic("Couldn't run " + app + " for " + action + " : " + err.Error())
	}

	app_name := app
	title := o.Name

	d.RenderArgs["title"] = title
	d.RenderArgs["app_name"] = app_name
	d.RenderArgs["data"] = data

	return d.RenderTemplate("Loader/LoadApp.html")
}

func (d Loader) SendMessage(app string) revel.Result {
	var toAddr []string
	var components []*models.Component
	for key, value := range d.Params.Form {
		if key == "to" {
			// To Address
			toAddr = value
		} else {
			components = append(components,
				&models.Component{
					Name: key,
					Data: []byte(value[0]),
				})
		}
	}

	u, err := d.Txn.Get(&models.User{}, GetUserId(d.Session))
	if err != nil {
		panic(err)
	}

	fromSub, err := models.UserIdentities(u.(*models.User), d.Txn)
	if err != nil {
		panic(err)
	}

	msg, err := models.CreateMessage(d.Txn, fromSub[0].Address.String(), toAddr, components)
	if err != nil {
		panic(err)
	}

	if toAddr != nil {
		mailserver.InitRouter()
		for _, addr := range toAddr {
			err := mailserver.SendAlert(mailserver.LookupRouter, msg.Name, fromSub[0], addr)
			if err != nil {
				panic(err)
			}
		}
	}

	return d.Redirect(routes.Loader.LoadAppDefault(app))
}
