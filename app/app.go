package app

import (
	"fmt"
	"net/http"

	"getmelange.com/app/framework"
)

func (r *Server) HandleApp(res http.ResponseWriter, req *http.Request) {
	// Serve Application Files
	view := framework.ServeFile("client", req.URL.Path)
	framework.WriteView(&framework.CSPWrapper{
		CSP: fmt.Sprintf("default-src 'self' %[1]s;"+
			"img-src *;"+
			"script-src 'self' %[1]s 'unsafe-eval';"+
			"frame-src 'self' %[2]s;"+
			"style-src 'self' %[1]s 'unsafe-inline';"+
			"connect-src 'self' %[1]s %[3]s %[4]s ws://api.melange.127.0.0.1.xip.io:7776;"+
			"font-src 'self' %[1]s;", r.CommonURL(), r.PluginURL(), r.APIURL(), r.APIRealtimeURL()),
		View: view,
	}, res)
}