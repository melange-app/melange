package app

import (
	"fmt"
	"melange/app/framework"
	"melange/app/models"
	"melange/dispatcher"
	"melange/tracker"
	"net/http"
	"strings"
)

type Server struct {
	Suffix  string
	Common  string
	Plugins string
	App     string
	API     string
	// Other Servers
	Dispatcher *dispatcher.Server
	Tracker    *tracker.Tracker
	// Settings Module
	Settings *models.Store
}

func (p *Server) CommonURL() string {
	return p.Common + p.Suffix
}

func (p *Server) PluginURL() string {
	return p.Plugins + p.Suffix
}

func (p *Server) APIURL() string {
	return p.API + p.Suffix
}

func (p *Server) AppURL() string {
	return p.App + p.Suffix
}

func (p *Server) Run(port int) error {
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: &Router{p},
	}
	return s.ListenAndServe()
}

type Router struct {
	p *Server
}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// fmt.Println("Got request for", req.URL.Path)
	// Ensure that the Host matches what we expect
	url := strings.Split(req.Host, ".melange")
	if len(url) != 2 || !(strings.HasPrefix(url[1], ":") || url[1] == r.p.Suffix) {
		framework.WriteView(framework.Error403, res)
		return
	}

	if req.URL.Path == "/favicon.ico" {
		framework.WriteView(framework.Error404, res)
		return
	}

	mode := url[0]

	if strings.HasSuffix(mode, "plugins") {
		r.p.HandlePlugins(mode, res, req)
	} else if mode == "common" {
		r.p.HandleCommon(res, req)
	} else if mode == "app" {
		r.p.HandleApp(res, req)
	} else if mode == "api" {
		r.p.HandleAPI(res, req)
	}
}
