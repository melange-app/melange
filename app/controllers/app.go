package controllers

import (
	"github.com/robfig/revel"
	"melange/app/models"
	"melange/app/routes"
	"melange/mailserver"
	"strconv"
)

func GetUserId(session map[string]string) int {
	i, _ := strconv.Atoi(session["user"])
	return i
}

type App struct {
	GorpController
}

func (c App) Index() revel.Result {
	_, ok := c.Session["user"]
	if ok {
		return c.Redirect(routes.Dispatch.Dashboard())
	}
	return c.Render()
}

// This action controls Logout
func (c App) Logout() revel.Result {
	delete(c.Session, "user")
	c.Flash.Success("You have successfully logged out.")
	return c.Redirect(routes.App.Index())
}

// These Two Actions Control Displaying the Login View and Processing Logins
func (c App) Login() revel.Result {
	_, ok := c.Session["user"]
	if ok {
		return c.Redirect(routes.Dispatch.Dashboard())
	}
	return c.Render()
}

func (c App) ProcessLogin(username string, password string) revel.Result {
	user, err := models.AuthenticateUser(username, password, c.Txn)
	if err != nil {
		c.Flash.Error("Unable to authenticate user.")
		return c.Redirect(routes.App.Login())
	}

	// User is logged in, redirect to dashboard.
	c.Session["user"] = strconv.Itoa(user.UserId)
	c.Flash.Success("Welcome back, " + user.Name + ".")
	return c.Redirect(routes.Dispatch.Dashboard())
}

// These Two Actions Control Displaying the Registration View and Processing Registration
func (c App) Register() revel.Result {
	return c.Render()
}

func (c App) ProcessRegistration(
	name string,
	username string,
	password string,
	confirmPassword string) revel.Result {

	// Validations Time
	c.Validation.Required(name).Message("Full Name is required.")
	c.Validation.MaxSize(name, 100).Message("Full name cannot be longer than 100 characters.")
	c.Validation.Required(username).Message("Username is required.")
	c.Validation.MaxSize(username, 20).Message("Username cannot be longer than 20 characters")
	c.Validation.Required(password).Message("Password is required.")
	c.Validation.Required(confirmPassword).Message("Confirmation password is required.")
	c.Validation.Required(password == confirmPassword).Message("Passwords must match.")

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.ProcessRegistration)
	}

	// Validation Passed, Let's Create the User
	u := models.CreateUser(username, password, name)
	u.Save(c.Txn)

	id, err := models.NewIdentityForUser(u, u.Username)
	if err != nil {
		panic(err)
	}
	c.Txn.Insert(id)

	mailserver.InitRouter()

	err = id.Register(mailserver.RegistrationRouter)
	if err != nil {
		panic(err)
	}

	c.Session["user"] = strconv.Itoa(u.UserId)
	c.Flash.Success("Hi there, " + u.Name + " looks like it's your first time here. Why not try the tutorial?")

	return c.Redirect(routes.Dispatch.Dashboard())
}
