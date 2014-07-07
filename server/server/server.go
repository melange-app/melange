package main

import (
	"melange"
	"melange/dispatcher"
	"melange/plugins"
	"melange/tracker"
)

// Main Server
type Melange struct {
	Plugins    plugins.Server
	Dispatcher dispatcher.Server
	Tracker    tracker.Server
}

func (m *Melange) Post(f *melange.Server, msg *melange.Message) {
	// if f == m.Plugins {
	// 	// Start the Other Server
	// }
}

func (m *Melange) Run(port int) error {
	m.Plugins = plugins.Server{
		Suffix:   ".127.0.0.1.xip.io:9001",
		Common:   "http://common.melange",
		Plugins:  "http://*.plugins.melange",
		Delegate: m,
	}
	return m.Plugins.Run(port)
}

func main() {
	mel := &Melange{}
	mel.Run(9001)
}
