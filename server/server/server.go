package main

import (
	"fmt"
	"melange/app"
	"melange/app/models"
	"melange/dispatcher"
	"melange/tracker"
	"net/http"
	"os"
	"path/filepath"
)

// Main Server
type Melange struct {
	App        *app.Server
	Dispatcher *dispatcher.Server
	Tracker    *tracker.Tracker
}

func (m *Melange) Run(port int) error {
	// Create a New Store for Settings
	dataDir := os.Getenv("MLGDATA")
	err := os.MkdirAll(dataDir, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	settings, err := models.CreateStore(filepath.Join(dataDir, "settings.db"))
	if err != nil {
		return err
	}

	m.App = &app.Server{
		Suffix:  ":7776",
		Common:  "http://common.melange",
		Plugins: "http://*.plugins.melange",
		API:     "http://api.melange",
		App:     "http://app.melange",
		// Other Servers
		Dispatcher: m.Dispatcher,
		Tracker:    m.Tracker,
		// Settings
		Settings: settings,
	}

	fmt.Println("Starting up.")
	go func() {
		port := os.Getenv("MLGPORT")
		if port != "" {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%s/startup", port))
			if err != nil || resp.StatusCode != 200 {
				fmt.Println("Got error getting startup...", err, resp.StatusCode)
			}
		}
	}()

	return m.App.Run(port)
}

func main() {
	mel := &Melange{}
	err := mel.Run(7776)
	if err != nil {
		fmt.Println("Error Starting Server", err)
	}
}
