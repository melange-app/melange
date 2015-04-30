package app

import (
	"fmt"
	"net/http"
	"path/filepath"

	"getmelange.com/app/framework"
)

func (r *Server) HandleApp(res http.ResponseWriter, req *http.Request) {
	// Serve Application Files
	clientBase := filepath.Join(r.AssetDirectory, "client")
	view := framework.ServeFile(clientBase, req.URL.Path)
	framework.WriteView(&framework.CSPWrapper{
		CSP: fmt.Sprintf("default-src 'self' %[1]s;"+
			"img-src *;"+
			"script-src 'self' %[1]s 'unsafe-eval';"+
			"frame-src 'self' %[2]s;"+
			"style-src 'self' %[1]s %[5]s 'unsafe-inline';"+
			"connect-src 'self' %[1]s %[3]s %[4]s ws://api.local.getmelange.com:7776;"+
			"font-src 'self' %[1]s %[5]s;", r.CommonURL(), r.PluginURL(), r.APIURL(), r.APIRealtimeURL(), "localhost:7776"),
		View: view,
	}, res)
}
