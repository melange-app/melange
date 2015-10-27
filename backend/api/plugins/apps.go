package plugins

import (
	"fmt"
	"net/http"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/packaging"
)

type appRequest struct {
	Repository string
	Id         string
}

type CheckForPluginUpdatesController struct {
	Packager *packaging.Packager
}

func (u *CheckForPluginUpdatesController) Handle(req *http.Request) framework.View {
	updates, err := u.Packager.CheckForPluginUpdates()
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

type UpdatePluginController struct {
	Packager *packaging.Packager
}

func (u *UpdatePluginController) Handle(req *http.Request) framework.View {
	r := &packaging.PluginUpdate{}
	err := DecodeJSONBody(req, r)
	if err != nil {
		fmt.Println("Couldn't deocde JSON Body", err)
		return framework.Error500
	}

	err = u.Packager.ExecuteUpdate(r)
	if err != nil {
		fmt.Println("Error updating plugin", err)
		return framework.Error500
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "OK",
	}
}

// InstallPluginController will install a plugin from a specific github repository.
type InstallPluginController struct {
	Packager *packaging.Packager
}

// Handle performs the installation.
func (i *InstallPluginController) Handle(req *http.Request) framework.View {
	r := &appRequest{}
	err := DecodeJSONBody(req, r)
	if err != nil {
		fmt.Println("Couldn't deocde JSON Body", err)
		return framework.Error500
	}

	err = i.Packager.InstallPlugin(r.Repository)
	if err != nil {
		fmt.Println("Error install application", err)
		return framework.Error500
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "OK",
	}
}

// UninstallPluginController uninstalls a plugin given the ID of the plugin.
type UninstallPluginController struct {
	Packager *packaging.Packager
}

// Handle performs the uninstallation.
func (i *UninstallPluginController) Handle(req *http.Request) framework.View {
	r := &appRequest{}
	err := DecodeJSONBody(req, r)
	if err != nil {
		fmt.Println("Couldn't deocde JSON Body", err)
		return framework.Error500
	}

	err = i.Packager.UninstallPlugin(r.Id)
	if err != nil {
		fmt.Println("Error uninstalling application", err)
		return framework.Error500
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "OK",
	}
}

// ServerLists will return the list of Trackers or Servers from
// getmelange.com.
type PluginStoreController struct {
	Packager *packaging.Packager
}

// Handle will decodeProviders from getmelange.com then return them in JSON.
func (s *PluginStoreController) Handle(req *http.Request) framework.View {
	packages, err := s.Packager.DecodeStore()
	if err != nil {
		fmt.Println("Getting store", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: packages,
	}
}
