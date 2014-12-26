package app

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"getmelange.com/app/framework"
)

func (r *Server) HandlePlugins(mode string, res http.ResponseWriter, req *http.Request) {
	// Serve Plugin Files
	pluginPath := strings.TrimSuffix(mode, ".plugins")
	view := framework.ServeFile(filepath.Join(r.DataDirectory, "plugins", pluginPath), req.URL.Path)
	framework.WriteView(&framework.CSPWrapper{
		CSP: fmt.Sprintf("default-src 'self';"+
			"img-src *;"+
			"font-src 'self' %[1]s;"+
			"script-src 'self' 'unsafe-eval' %[1]s;"+
			"style-src 'self' 'unsafe-inline' %[1]s", r.CommonURL()),
		View: view,
	}, res)
}
