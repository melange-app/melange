package app

import (
	"net/http"
	"path/filepath"
	"strings"

	"getmelange.com/app/framework"
)

type CORSFile struct {
	framework.View
}

// Headers add the CORS headers to the request:
//
// "Access-Control-Allow-Origin"
// "Access-Control-Allow-Headers"
func (a *CORSFile) Headers() framework.Headers {
	hdrs := make(framework.Headers)
	hdrs["Access-Control-Allow-Origin"] = "*"
	hdrs["Access-Control-Allow-Headers"] = "Content-Type"
	return hdrs
}

func (r *Server) HandleCommon(res http.ResponseWriter, req *http.Request) {
	libBase := filepath.Join(r.AssetDirectory, "lib")

	// fmt.Println("Get Common", req.URL.Path)
	// Serve Library Files
	if req.URL.Path == "/main/theme" {
		// Load the Main Theme Files
		framework.WriteView(framework.ServeFile(libBase, filepath.Join("bootswatch-yeti", "3.1.1.css")), res)
	} else {
		dirs := strings.Split(req.URL.Path, "/")
		// No More Panics
		if len(dirs) != 4 {
			framework.WriteView(framework.Error404, res)
			return
		}
		typ, lib, version := dirs[1], dirs[2], dirs[3]
		// fmt.Println("About to serve file", typ, lib, version)

		view := framework.ServeFile(libBase, filepath.Join(filepath.FromSlash(lib), version+"."+filepath.FromSlash(typ)))

		// Resources from Common can be Accessed by All Origins
		framework.WriteView(
			&CORSFile{
				View: view,
			}, res)
	}
}
