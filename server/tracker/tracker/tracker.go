package main

import (
	"flag"
	"fmt"
	"melange/tracker"
)

// Start the Tracker
func main() {
	var (
		keyFile  = flag.String("keyfile", "", "Location of file to get keys from.")
		saveFile = flag.String("savefile", "", "Location of file to save requests to.")
		port     = flag.Int("port", 1024, "Port to run server on.")
	)
	flag.Parse()

	mel := &tracker.Tracker{
		KeyFile:  *keyFile,
		SaveFile: *saveFile,
	}
	err := mel.Run(*port)
	if err != nil {
		fmt.Println("Error Starting Server", err)
	}
}
