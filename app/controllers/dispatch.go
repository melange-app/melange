package controllers

import (
	"github.com/airdispatch/dpl"
	"github.com/robfig/revel"
	"melange/app/models"
	"melange/app/routes"
	"net/http"
)

type Dispatch struct {
	GorpController
}

func (c Dispatch) Init() revel.Result {
	if c.Session["user"] == "" {
		return c.Redirect(routes.App.Login())
	}

	// Function to load user apps
	apps, err := models.GetUserApps(c.Txn, c.Session)
	if err != nil {
		panic(err)
	}

	c.RenderArgs["apps"] = apps

	return nil
}

func (d Dispatch) Dashboard() revel.Result {
	// Download all recents from subscribed, download all recents from alerts, sort them chronologically
	recents := []int{0}
	return d.Render(recents)
}

func (d Dispatch) Profile() revel.Result {
	return d.Render()
}

func (d Dispatch) All() revel.Result {
	return d.Todo()
}

func (d Dispatch) Applications() revel.Result {
	return d.Render()
}

func (d Dispatch) AddApplication(url string) revel.Result {
	resp, err := http.Get(url)
	if err != nil {
		d.Flash.Error("Unable to download application via supplied url.")
		return d.Redirect(routes.Dispatch.Applications())
	}
	defer resp.Body.Close()

	plugin, err := dpl.ParseDPLStream(resp.Body)
	if err != nil {
		d.Flash.Error("Plugin is not valid.")
		return d.Redirect(routes.Dispatch.Applications())
	}

	in := &models.UserApp{
		UserId: GetUserId(d.Session),
		AppURL: url,
		Name:   plugin.Name,
		Path:   plugin.Path,
	}
	err = d.Txn.Insert(in)
	if err != nil {
		panic(err)
	}

	return d.Redirect(routes.Dispatch.Applications())
}

func (d Dispatch) UninstallApplication(app string) revel.Result {
	u := GetUserId(d.Session)

	var apps []*models.UserApp
	_, err := d.Txn.Select(&apps, "select * from dispatch_app where userid = $1 and UPPER(name) = UPPER($2)", u, app)
	if err != nil {
		panic(err)
	}
	if len(apps) != 1 {
		panic(len(apps))
	}

	toDelete := apps[0]
	count, err := d.Txn.Delete(toDelete)
	if err != nil || count != 1 {
		panic(err)
	}

	return d.Redirect(routes.Dispatch.Applications())
}

func (c Dispatch) Account() revel.Result {
	if c.Session["user"] == "" {
		return c.Redirect(routes.App.Login())
	}
	var users []*models.User

	_, err := c.Txn.Select(&users, "select * from dispatch_user where userid = $1", GetUserId(c.Session))
	if err != nil {
		panic(err)
	}
	user := users[0]

	var subscriptions []*models.UserSubscription
	_, err = c.Txn.Select(&subscriptions, "select * from dispatch_subscription where userid = $1", user.UserId)
	if err != nil {
		panic(err)
	}

	return c.Render(user, subscriptions)
}

func (c Dispatch) AddSubscription(address string) revel.Result {
	// TODO: Verify the Address Somehow...
	if false {
		c.Flash.Error("Unable to verify that address. It has not been added to your subscriptions.")
		return c.Redirect(routes.Dispatch.Account())
	}

	user := GetUserId(c.Session)
	newSub := &models.UserSubscription{
		UserId:  user,
		Address: address,
	}

	err := c.Txn.Insert(newSub)
	if err != nil {
		panic(err)
	}

	return c.Redirect(routes.Dispatch.Account())
}

func (d Dispatch) RemoveSubscription(id int) revel.Result {
	var apps []*models.UserSubscription
	_, err := d.Txn.Select(&apps, "select * from dispatch_subscription where subscriptionid = $1", id)
	if err != nil {
		panic(err)
	}
	if len(apps) != 1 {
		panic(len(apps))
	}

	toDelete := apps[0]
	count, err := d.Txn.Delete(toDelete)
	if err != nil || count != 1 {
		panic(err)
	}

	return d.Redirect(routes.Dispatch.Account())
}

func (c Dispatch) ProcessAccount() revel.Result {
	return c.Render()
}
