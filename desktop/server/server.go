// This file provides a way to parse the Melange server information
// from the environmental variables specified when launching the
// program. This is generally done through the electron main.js file.
package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"getmelange.com/backend"
	"getmelange.com/backend/info"
	"getmelange.com/backend/models"
	"getmelange.com/dispatcher"
	"getmelange.com/tracker"
)

const (
	portNumber = 7776
	suffix     = ".local.getmelange.com:%d"
)

// Main Server
type Melange struct {
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

	store, err := models.CreateStore(filepath.Join(dataDir, "settings.db"))
	if err != nil {
		return err
	}

	environment := &info.Environment{
		Settings: info.Settings{
			Suffix: fmt.Sprintf(suffix, portNumber),

			Version:  os.Getenv("MLGVERSION"),
			Platform: os.Getenv("MLGPLATFORM"),

			// Settings
			Store: store,

			// Logging Information
			DataDirectory: dataDir,

			ControllerPort: os.Getenv("MLGPORT"),
			AppLocation:    os.Getenv("MLGAPP"),

			Debug: os.Getenv("MLGDEBUG") != "",
		},
	}

	fmt.Println("Starting up.")
	go func() {
		port := environment.ControllerPort
		if port != "" {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%s/startup", port))
			if err != nil || resp.StatusCode != 200 {
				fmt.Println("Got error getting startup...", err)
			}
		}
	}()

	// Start the Melange server
	return backend.Start(environment, portNumber)
}

func main() {
	mel := &Melange{}
	err := mel.Run(7776)
	if err != nil {
		fmt.Println("Error Starting Server", err)
	}
}
