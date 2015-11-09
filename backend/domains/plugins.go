package domains

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/info"
)

func HandlePlugins(
	mode string,
	res http.ResponseWriter,
	req *http.Request,
	env *info.Environment,
) {
	pluginID := strings.TrimSuffix(mode, ".plugins")
	pluginPath := filepath.Join(env.DataDirectory, "plugins", pluginID)

	view := framework.ServeFile(pluginPath, req.URL.Path)

	framework.WriteView(&framework.CSPWrapper{
		CSP: fmt.Sprintf("default-src 'self';"+
			"img-src *;"+
			"font-src 'self' %[1]s;"+
			"script-src 'self' 'unsafe-eval' %[1]s;"+
			"style-src 'self' 'unsafe-inline' %[1]s", env.CommonURL()),
		View: view,
	}, res)
}
