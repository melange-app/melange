package main

import (
	"flag"
	"fmt"
	"getmelange.com/tracker"
)

// Start the Tracker
func main() {
	var (
		keyFile = flag.String("keyfile", "", "Location of file to get keys from.")
		dbConn  = flag.String("dbConn", "", "DB Connection String")
		port    = flag.Int("port", 1024, "Port to run server on.")
	)
	flag.Parse()

	mel := &tracker.Tracker{
		KeyFile:  *keyFile,
		DBString: *dbConn,
	}
	err := mel.Run(*port)
	if err != nil {
		fmt.Println("Error Starting Server", err)
	}
}
