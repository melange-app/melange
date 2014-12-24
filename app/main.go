package app

import (
	"fmt"
	"net/http"
	"strings"

	"getmelange.com/app/framework"
	"getmelange.com/app/models"
	"getmelange.com/dispatcher"
	"getmelange.com/tracker"
)

const webProtocol = "http://"
const webSocketProtocol = "ws://"

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
	return webProtocol + p.Common + p.Suffix
}

func (p *Server) PluginURL() string {
	return webProtocol + p.Plugins + p.Suffix
}

func (p *Server) APIURL() string {
	return webProtocol + p.API + p.Suffix
}

func (p *Server) AppURL() string {
	return webProtocol + p.App + p.Suffix
}

func (p *Server) APIRealtimeURL() string {
	return webSocketProtocol + p.API + p.Suffix
}

func (p *Server) Run(port int) error {
	s := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: &Router{p},
	}

	fmt.Println("Running HTTP Server")
	return s.ListenAndServe()
}

type Router struct {
	p *Server
}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// Ensure that the Host matches what we expect
	url := strings.Split(req.Host, ".melange")

	if len(url) > 2 {
		url = []string{
			strings.Join(url[:len(url)-1], ".melange"),
			url[len(url)-1],
		}
	}

	// if (len(url) != 2 || !(strings.HasPrefix(url[1], ":") || url[1] == r.p.Suffix)) &&
	// 	(req.URL.Path != "/realtime") && (url[0] != "data") {
	// 	framework.WriteView(framework.Error403, res)
	// 	return
	// }

	if req.URL.Path == "/favicon.ico" {
		framework.WriteView(framework.Error404, res)
		return
	}

	mode := url[0]

	fmt.Println(req.Method, req.URL.Path, "on", req.Host, mode)

	if strings.HasSuffix(mode, "plugins") {
		r.p.HandlePlugins(mode, res, req)
	} else if mode == "common" {
		r.p.HandleCommon(res, req)
	} else if mode == "app" {
		r.p.HandleApp(res, req)
	} else if mode == "api" {
		r.p.HandleAPI(res, req)
	} else if mode == "data" {
		r.p.HandleData(res, req)
	}
}
