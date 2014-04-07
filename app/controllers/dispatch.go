package controllers

import (
	"github.com/airdispatch/dpl"
	"github.com/robfig/revel"
	"melange/app/models"
	"melange/app/routes"
	"melange/mailserver"
	"net/http"
	"time"
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

	mailserver.InitRouter()

	return nil
}

func (d Dispatch) Dashboard() revel.Result {
	// Download all recents from subscribed, download all recents from alerts, sort them chronologically
	u, err := d.Txn.Get(&models.User{}, GetUserId(d.Session))
	if err != nil {
		panic(err)
	}

	id, err := models.UserIdentities(u.(*models.User), d.Txn)
	if err != nil {
		panic(err)
	}
	if len(id) == 0 {
		panic("Not enough IDs")
	}

	recents, err := mailserver.Messages(mailserver.LookupRouter,
		d.Txn,
		id[0],
		u.(*models.User),
		true, true, true,
		time.Now().Add(-7*24*time.Hour).Unix())
	if err != nil {
		panic(err)
	}

	return d.Render(recents)
}

func (d Dispatch) Profile() revel.Result {
	user, err := d.Txn.Get(&models.User{}, GetUserId(d.Session))
	if err != nil {
		panic(err)
	}

	id, err := models.UserIdentities(user.(*models.User), d.Txn)
	if err != nil {
		panic(err)
	}

	messages, err := mailserver.Messages(mailserver.LookupRouter,
		d.Txn,
		id[0],
		user.(*models.User),
		false, false, true,
		time.Now().Add(-7*24*time.Hour).Unix())
	if err != nil {
		panic(err)
	}

	return d.Render(messages, user)
}

func (d Dispatch) All() revel.Result {
	return d.Render()
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

	return d.Redirect(routes.Loader.LoadAppDefault(plugin.Name))
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

	var identities []*models.Identity
	_, err = c.Txn.Select(&identities, "select * from dispatch_identity where userid = $1", user.UserId)
	if err != nil {
		panic(err)
	}

	return c.Render(user, subscriptions, identities)
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
	u := GetUserId(d.Session)

	var apps []*models.UserSubscription
	_, err := d.Txn.Select(&apps, "select * from dispatch_subscription where subscriptionid = $1 and userid = $2", id, u)
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

func (d Dispatch) RegisterIdentity(id int) revel.Result {
	u := GetUserId(d.Session)

	var ids []*models.Identity
	_, err := d.Txn.Select(&ids, "select * from dispatch_identity where identityid = $1 and userid = $2", id, u)
	if err != nil {
		panic(err)
	}
	if len(ids) != 1 {
		panic(len(ids))
	}

	toRegister := ids[0]

	mailserver.InitRouter()
	did, err := toRegister.ToDispatch()
	if err != nil {
		panic(err)
	}
	did.SetLocation(revel.Config.StringDefault("server.location", ""))

	err = mailserver.RegistrationRouter.Register(did, toRegister.Alias)
	if err != nil {
		panic(err)
	}

	return d.Redirect(routes.Dispatch.Account())
}

func (d Dispatch) ProcessAccount(name string, username string, password1 string, password2 string, password string) revel.Result {
	u, err := d.Txn.Get(&models.User{}, GetUserId(d.Session))
	if err != nil {
		panic(err)
	}

	user := u.(*models.User)
	user.Name = name

	if user.VerifyPassword(password) {
		if password1 == password2 {
			user.UpdatePassword(password1)
		} else {
			d.Flash.Error("New passwords do not match.")
		}
	} else {
		if password != "" {
			d.Flash.Error("Current Password is not correct. Did not update username or password.")
		}
	}

	user.Save(d.Txn)
	return d.Redirect(routes.Dispatch.Account())
}
