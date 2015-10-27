// Package info provides information about the currently running
// Melange backend to the Controllers that need it.
package info

import (
	"melange/app/models"
	"path/filepath"

	"getmelange.com/backend/packaging"

	gdb "github.com/huntaub/go-db"
)

const (
	CommonPrefix  = "common"
	PluginsPrefix = "plugins"
	AppPrefix     = "app"
	APIPrefix     = "api"
	DataPrefix    = "data"
)

const (
	webProtocol       = "http://"
	webSocketProtocol = "ws://"
)

// Environment represents information about the version of Melange
// that is currently running.
type Environment struct {
	// This is configurable based on how we are doing DNS loopback
	// to the device.
	Suffix string

	// Uncomment to include the server code that would power a
	// dispatcher and tracker.
	// Dispatcher *dispatcher.Server
	// Tracker *tracker.Tracker

	// Data storage
	Settings *models.Store

	// Local directories used for locating plugins and assets.
	DataDirectory  string
	AssetDirectory string

	// Information about the device that we are running on.
	Platform string
	Version  string

	// Information about how Melange is being run.
	AppLocation    string
	ControllerPort string

	Debug bool

	cachedTables map[string]gdb.Table
}

// CommonURL returns the url of the common subdomain.
func (e *Environment) CommonURL() string {
	return webProtocol + CommonPrefix + e.Suffix
}

// PluginURL returns the url of the plugin subdomain.
func (e *Environment) PluginURL() string {
	return webProtocol + PluginsPrefix + e.Suffix
}

// APIURL returns the url of the API subdomain.
func (e *Environment) APIURL() string {
	return webProtocol + APIPrefix + e.Suffix
}

// AppURL returns the url of the app subdomain.
func (e *Environment) AppURL() string {
	return webProtocol + AppPrefix + e.Suffix
}

// APIRealtimeURL returns the hostname of the realtime url
func (e *Environment) APIRealtimeURL() string {
	return webSocketProtocol + apiPrefix + e.Suffix
}

// Tables returns a map of all of all of the ORM database tables.
func (e *Environment) Tables() (map[string]gdb.Table, error) {
	// We only want to create the tables once per
	// server-invocation.
	if e.cachedTables == nil {
		e.cachedTables = models.CreateTables(e.Settings)
	}

	return e.cachedTables
}

func (e *Environment) Packager() *packaging.Packager {
	packager := &packaging.Packager{
		API:    "http://www.getmelange.com/api",
		Plugin: filepath.Join(e.DataDirectory, "plugins"),
		Debug:  e.Debug,
	}
	packager.CreatePluginDirectory()

	return packager
}
