package controllers

import (
	"fmt"
	"github.com/airdispatch/dpl"
	"github.com/robfig/revel"
	"io/ioutil"
	//"melange/app/routes"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Loader struct {
	Dispatch
}

type MessageTest struct {
	Num int
}

func (m *MessageTest) Get(field string) ([]byte, error) {
	return []byte(field + " " + strconv.Itoa(m.Num)), nil
}

func (m *MessageTest) Has(field string) bool {
	return true
}

func (m *MessageTest) Created() time.Time {
	return time.Now()
}

func (m *MessageTest) Sender() dpl.User {
	url, _ := url.Parse("https://fbcdn-sphotos-f-a.akamaihd.net/hphotos-ak-ash4/302066_10151616869615972_963571219_n.jpg")
	return dpl.User{
		Name:   "Hunter Leath",
		Avatar: url,
	}
}

type PluginHost struct{}

func (p *PluginHost) GetMessages(plugin *dpl.PluginInstance, tag dpl.Tag, predicate *dpl.Predicate, limit int) ([]dpl.Message, error) {
	messages := make([]dpl.Message, 0)
	for i := 0; i < 10; i++ {
		messages = append(messages, &MessageTest{i})
	}
	return messages, nil
}

func (p *PluginHost) GetURLForAction(
	plugin *dpl.PluginInstance, // The Plugin Calling for the URL
	action dpl.Action, // The Action
	message dpl.Message, // MessageContext
	user *dpl.User) (*url.URL, error) { // User Context

	rawURL := fmt.Sprintf("/app/%s/%s", strings.ToLower(plugin.Name), action.Name)
	if message != nil {
		rawURL = fmt.Sprintf("%s/m/%d", rawURL, message.(*MessageTest).Num)
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

var GlobalHost *PluginHost

func (d Loader) LoadAppMessage(app string, action string, message int) revel.Result {
	d.SetAction("Loader", "LoadApp")
	return d.LoadApp(app, action, &MessageTest{}, nil)
}

func (d Loader) LoadAppUser(app string, action string, username string) revel.Result {
	d.SetAction("Loader", "LoadApp")
	return d.LoadApp(app, action, nil, &dpl.User{})
}

// We should load the Application just once from the XML, then refresh it at a regular interval
func (d Loader) LoadApp(app string, action string, message dpl.Message, user *dpl.User) revel.Result {
	xmlFile, err := os.Open("../github.com/airdispatch/dpl/example/" + app + ".dpl")
	if err != nil {
		panic("Couldn't find App")
	}
	defer xmlFile.Close()

	b, _ := ioutil.ReadAll(xmlFile)

	// The Plugin is O
	o, err := dpl.ParseDPL(b)
	if err != nil {
		panic("Couldn't Parse App: " + err.Error())
	}

	// Create a Singleton that Hosts all the Plugins
	if GlobalHost == nil {
		GlobalHost = &PluginHost{}
	}

	plugin := o.CreateInstance(GlobalHost, nil)

	data, err := plugin.RunActionWithContext(action, message, user)
	if err != nil {
		panic("Couldn't run " + app + " for " + action + " : " + err.Error())
	}

	app_name := app
	title := o.Name
	return d.Render(title, app_name, data)
}

func (d Loader) AppPost(app string, action string) revel.Result {
	return d.Todo()
}

func (d Loader) SendMessage() revel.Result {
	return d.Todo()
}
