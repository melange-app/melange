package installer

import "getmelange.com/backend/api/router"

// UpdateRouter holds all of the routes associated with the update api
// module.
var UpdateRouter = router.CreateRouter("/updates").
	AddRoute("/download/progress", &UpdateProgressController{}).
	AddRoute("/download", &DownloadUpdateController{}).
	AddRoute("/install", &InstallUpdateController{}).
	AddRoute("", &CheckUpdateController{})

// Router holds all of the routes associated with the plugins api
// module.
var Router = router.CreateRouter("/plugins").
	AddRoute("/install", &InstallPluginController{}).
	AddRoute("/uninstall", &UninstallPluginController{}).
	AddRoute("/updates", &CheckForPluginUpdatesController{}).
	AddRoute("/update", &UpdatePluginController{})
