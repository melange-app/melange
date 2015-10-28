package backend

import (
	"fmt"
	"net/http"
	"strings"

	"getmelange.com/backend/api"
	"getmelange.com/backend/domains"
	"getmelange.com/backend/framework"
	"getmelange.com/backend/info"
)

func Start(info *info.Environment, port int) error {
	r := &Router{
		Environment: Environment,
	}

	s := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: &Router{p},
	}

	if err := info.Cache(); err != nil {
		return err
	}

	fmt.Println("Running HTTP Server")
	return s.ListenAndServe()
}

type Router struct {
	*info.Environment
}

func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// Ensure that the Host matches what we expect
	url := strings.Split(req.Host, s.Environment.Suffix)

	if len(url) > 2 {
		url = []string{
			strings.Join(url[:len(url)-1], s.Environment.Suffix),
			url[len(url)-1],
		}
	}

	// Return a 403 error if the URL does not match the expected
	// patterns.
	if (len(url) != 2 || !(strings.HasPrefix(url[1], ":") || url[1] == r.p.Suffix)) &&
		(req.URL.Path != "/realtime") &&
		(url[0] != "data") {
		framework.WriteView(framework.Error403, res)
		return
	}

	// No need to serve a favicon on any of the domains.
	if req.URL.Path == "/favicon.ico" {
		framework.WriteView(framework.Error404, res)
		return
	}

	mode := url[0]

	// Log every request that we handle.
	if s.Environment.Debug {
		fmt.Println(req.Method, req.URL.Path, "on", req.Host, mode)
	}

	env := s.Environment

	// Delegate to the correct domain handler.
	switch {
	case strings.HasSuffix(mode, environment.PluginsPrefix):
		domains.HandlePlugins(mode, res, req, env)
	case mode == environment.CommonPrefix:
		domains.HandleCommon(res, req, env)
	case mode == environment.AppPrefix:
		domains.HandleApp(res, req, env)
	case mode == environment.DataPrefix:
		domains.HandleData(res, req, env)
	case mode == environment.APIPrefix:
		api.HandleRequest(res, req, env)
	}

}
