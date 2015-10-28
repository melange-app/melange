package info

import "getmelange.com/backend/models"

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

// Settings stores platform-specific information about the Melange
// server that is currently running.
type Settings struct {
	// This is configurable based on how we are doing DNS loopback
	// to the device.
	Suffix string

	// Uncomment to include the server code that would power a
	// dispatcher and tracker.
	// Dispatcher *dispatcher.Server
	// Tracker *tracker.Tracker

	// Data storage
	Store *models.Store

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
}

// CommonURL returns the url of the common subdomain.
func (e Settings) CommonURL() string {
	return webProtocol + CommonPrefix + e.Suffix
}

// PluginURL returns the url of the plugin subdomain.
func (e Settings) PluginURL() string {
	return webProtocol + PluginsPrefix + e.Suffix
}

// APIURL returns the url of the API subdomain.
func (e Settings) APIURL() string {
	return webProtocol + APIPrefix + e.Suffix
}

// AppURL returns the url of the app subdomain.
func (e Settings) AppURL() string {
	return webProtocol + AppPrefix + e.Suffix
}

// APIRealtimeURL returns the hostname of the realtime url
func (e Settings) APIRealtimeURL() string {
	return webSocketProtocol + APIPrefix + e.Suffix
}
