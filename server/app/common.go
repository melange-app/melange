package app

import (
	"net/http"
	"path/filepath"
	"strings"

	"getmelange.com/app/framework"
)

func (r *Server) HandleCommon(res http.ResponseWriter, req *http.Request) {
	// fmt.Println("Get Common", req.URL.Path)
	// Serve Library Files
	if req.URL.Path == "/main/theme" {
		// Load the Main Theme Files
		framework.WriteView(framework.ServeFile("lib", filepath.Join("bootswatch-yeti", "3.1.1.css")), res)
	} else {
		dirs := strings.Split(req.URL.Path, "/")
		// No More Panics
		if len(dirs) != 4 {
			framework.WriteView(framework.Error404, res)
			return
		}
		typ, lib, version := dirs[1], dirs[2], dirs[3]
		// fmt.Println("About to serve file", typ, lib, version)
		view := framework.ServeFile("lib", filepath.Join(filepath.FromSlash(lib), version+"."+filepath.FromSlash(typ)))
		framework.WriteView(view, res)
	}
}
