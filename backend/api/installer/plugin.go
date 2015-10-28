package installer

import (
	"fmt"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"
	"getmelange.com/backend/packaging"
)

type appRequest struct {
	Repository string
	Id         string
}

// CheckForPluginUpdatesController will check to see if any plugins
// have updates available.
type CheckForPluginUpdatesController struct{}

// Get will execute the check.
func (u *CheckForPluginUpdatesController) Get(req *router.Request) framework.View {
	updates, err := req.Environment.Packager.CheckForPluginUpdates()
	if err != nil {
		fmt.Println("Got error looking for plugin updates", err)
		return framework.Error500
	}

	if len(updates) == 0 {
		updates = make([]*packaging.PluginUpdate, 0)
	}

	return &framework.JSONView{
		Content: updates,
	}
}

// UpdatePluginController will update a plugin to a new version.
type UpdatePluginController struct{}

// Post will execute the update.
func (u *UpdatePluginController) Post(req *router.Request) framework.View {
	r := &packaging.PluginUpdate{}
	if err := req.JSON(r); err != nil {
		fmt.Println("Couldn't deocde JSON Body", err)
		return framework.Error500
	}

	// Actually install the plugin.
	if err := req.Environment.Packager.ExecuteUpdate(r); err != nil {
		fmt.Println("Error updating plugin", err)
		return framework.Error500
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "OK",
	}
}

// InstallPluginController will install a plugin from a specific github repository.
type InstallPluginController struct{}

// Handle performs the installation.
func (i *InstallPluginController) Post(req *router.Request) framework.View {
	r := &appRequest{}
	if err := req.JSON(r); err != nil {
		fmt.Println("Couldn't decode JSON Body", err)
		return framework.Error500
	}

	if err := req.Environment.Packager.InstallPlugin(r.Repository); err != nil {
		fmt.Println("Error install application", err)
		return framework.Error500
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "OK",
	}
}

// UninstallPluginController uninstalls a plugin given the ID of the plugin.
type UninstallPluginController struct{}

// Handle performs the uninstallation.
func (i *UninstallPluginController) Post(req *router.Request) framework.View {
	r := &appRequest{}
	if err := req.JSON(r); err != nil {
		fmt.Println("Couldn't decode JSON Body", err)
		return framework.Error500
	}

	if err := req.Environment.Packager.UninstallPlugin(r.Id); err != nil {
		fmt.Println("Error uninstalling application", err)
		return framework.Error500
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "OK",
	}
}
