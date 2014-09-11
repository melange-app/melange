package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"

	"getmelange.com/app"
	"getmelange.com/app/models"
	"getmelange.com/dispatcher"
	"getmelange.com/tracker"
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
		Common:  "common.melange",
		Plugins: "*.plugins.melange",
		API:     "api.melange",
		App:     "app.melange",
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
	var (
		cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	)
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for sig := range c {
				fmt.Println("Caught", sig)
				pprof.StopCPUProfile()
				os.Exit(0)
			}
		}()
	}

	mel := &Melange{}
	err := mel.Run(7776)
	if err != nil {
		fmt.Println("Error Starting Server", err)
	}
}
