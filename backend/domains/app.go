package domains

import (
	"fmt"
	"net/http"
	"path/filepath"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/info"
)

// HandleApp will return the requested file from the application
// domain to the client.
//
// The application domain handles serving the HTML, CSS, and JS
// included with Melange.
func HandleApp(
	res http.ResponseWriter,
	req *http.Request,
	env *info.Environment,
) {
	// Serve Application Files
	clientBase := filepath.Join(env.AssetDirectory, "client")
	view := framework.ServeFile(clientBase, req.URL.Path)

	framework.WriteView(&framework.CSPWrapper{
		CSP: fmt.Sprintf("default-src 'self' %[1]s;"+
			"img-src *;"+
			"script-src 'self' %[1]s 'unsafe-eval';"+
			"frame-src 'self' %[2]s;"+
			"style-src 'self' %[1]s %[5]s 'unsafe-inline';"+
			"connect-src 'self' %[1]s %[3]s %[4]s ws://api.local.getmelange.com:7776;"+
			"font-src 'self' %[1]s %[5]s;",

			env.CommonURL(),      // %[1]s
			env.PluginURL("*"),   // %[2]s
			env.APIURL(),         // %[3]s
			env.APIRealtimeURL(), // %[4]s
			"localhost:7776"),    // %[5]s
		View: view,
	}, res)
}
