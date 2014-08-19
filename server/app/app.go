package app

import (
	"fmt"
	"getmelange.com/app/framework"
	"net/http"
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
			"connect-src 'self' %[1]s %[3]s;"+
			"font-src 'self' %[1]s;", r.CommonURL(), r.PluginURL(), r.APIURL()),
		View: view,
	}, res)
}
