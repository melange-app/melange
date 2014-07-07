package app

import (
	"fmt"
	"melange/app/framework"
	"net/http"
	"path/filepath"
	"strings"
)

func (r *Server) HandlePlugins(mode string, res http.ResponseWriter, req *http.Request) {
	// Serve Plugin Files
	pluginPath := strings.TrimSuffix(mode, ".plugins")
	view := framework.ServeFile(filepath.Join("plugins", pluginPath), req.URL.Path)
	framework.WriteView(&framework.CSPWrapper{
		CSP: fmt.Sprintf("default-src 'self';"+
			"img-src *;"+
			"font-src 'self' %[1]s;"+
			"script-src 'self' 'unsafe-eval' %[1]s;"+
			"style-src 'self' 'unsafe-inline' %[1]s", r.CommonURL()),
		View: view,
	}, res)
}
