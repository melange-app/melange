package app

import (
	"fmt"
	"melange/app/framework"
	"net/http"
	"os"
	"path/filepath"
)

func (r *Server) HandleApp(res http.ResponseWriter, req *http.Request) {
	// Serve Application Files
	if req.URL.Path == "/plugins.json" {
		// Serve the Plugins.JSON Document
		framework.WriteView(framework.LoadPlugins(filepath.Join(os.Getenv("MLGBASE"), "plugins")), res)
	} else {
		// Serve Regular Files
		view := framework.ServeFile("client", req.URL.Path)
		framework.WriteView(&framework.CSPWrapper{
			CSP: fmt.Sprintf("default-src 'self' %[1]s;"+
				"img-src *;"+
				"script-src 'self' %[1]s 'unsafe-eval';"+
				"frame-src 'self' %[2]s;"+
				"style-src 'self' %[1]s 'unsafe-inline';"+
				"connect-src 'self' %[1]s;"+
				"font-src 'self' %[1]s;", r.CommonURL(), r.PluginURL()),
			View: view,
		}, res)
	}
}
