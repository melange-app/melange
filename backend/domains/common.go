package domains

import (
	"net/http"
	"path/filepath"
	"strings"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/info"
)

// HandleCommon will resolve URLs that occur on the Common domain.
//
// This will serve files out of the `lib` directory in
// `getmelange.com`.
func HandleCommon(
	res http.ResponseWriter,
	req *http.Request,
	env *info.Environment,
) {
	libBase := filepath.Join(env.AssetDirectory, "lib")

	var view framework.View

	// Serve Library Files
	if req.URL.Path == "/main/theme" {
		// Load the Main Theme Files
		mainTheme := filepath.Join("bootswatch-yeti", "3.3.4.css")
		view = framework.ServeFile(libBase, mainTheme)
	} else {
		dirs := strings.Split(req.URL.Path, "/")

		if len(dirs) != 4 {
			framework.WriteView(framework.Error404, res)
			return
		}
		typ, lib, version := dirs[1], dirs[2], dirs[3]

		location := filepath.Join(
			filepath.FromSlash(lib),
			version+"."+filepath.FromSlash(typ),
		)
		view = framework.ServeFile(libBase, location)
	}

	// Resources from Common can be Accessed by All Origins
	framework.WriteView(
		&framework.CORSWrapper{
			Origin: "*",
			View:   view,
		}, res)
}
