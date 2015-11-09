package melange

import (
	"fmt"
	"os"
	"path/filepath"

	"getmelange.com/app"
	"getmelange.com/app/models"
	"getmelange.com/dispatcher"
	"getmelange.com/tracker"
)

// Main Server
type melange struct {
	App        *app.Server
	Dispatcher *dispatcher.Server
	Tracker    *tracker.Tracker
}

func Run(port int, dataDir string, version string, platform string) error {
	return RunDarwin(port, dataDir, "", version, platform)
}

func RunDarwin(port int, dataDir string, assetDir string, version string, platform string) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Golang Panic:", r)
		}
	}()

	m := &melange{}

	// fmt.Println("Creating the data dir")
	// Create a New Store for Settings
	err := os.MkdirAll(dataDir, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	// fmt.Println("Creating the Database", dataDir)
	settings, err := models.CreateStore(filepath.Join(dataDir, "settings.db"))
	if err != nil {
		fmt.Println("Got error creating store", err)
		return err
	}

	suffix := fmt.Sprintf(":%d", port)
	if assetDir == "" {
		suffix = fmt.Sprintf(".127.0.0.1.xip.io:%d", port)
	}

	// fmt.Println("Creating the server")
	m.App = &app.Server{
		Suffix:  suffix,
		Common:  "common.melange",
		Plugins: "*.plugins.melange",
		API:     "api.melange",
		App:     "app.melange",
		// Other Servers
		// Dispatcher: m.Dispatcher,
		// Tracker:    m.Tracker,
		// Settings
		Settings: settings,

		// Logging Information
		DataDirectory:  dataDir,
		AssetDirectory: assetDir,
		Platform:       platform,
		Version:        version,
	}

	fmt.Println("Melange Starting up.", port)

	go func() {
		if err := m.App.Run(port); err != nil {
			fmt.Println("Encountered Error Running Application", err)
		}
	}()

	return nil
}
