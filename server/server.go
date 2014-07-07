package main

import (
	// "melange/mailserver"
	"melange/plugins"
	// "melange/tracker"
)

type Server interface {
	Run(port int) error
}

func main() {
	p := plugins.Server{
		Suffix:  ".127.0.0.1.xip.io:9001",
		Common:  "http://common.melange",
		Plugins: "http://*.plugins.melange",
	}
	p.Run(9001)
}
