package main

import (
	"flag"
	"fmt"
	"melange/dispatcher"
)

// Start the Dispatcher
func main() {
	var (
		me      = flag.String("me", "", "Location of the server.")
		keyFile = flag.String("keyfile", "", "Location of file to get keys from.")

		dbConn = flag.String("dbConn", "", "Connection string for Database.")
		dbType = flag.String("dbType", "", "Type of database. `psql`, `mysql`, `mssql`, `sqlite`.")

		port = flag.Int("port", 1024, "Port to run server on.")

		registerAt = flag.String("registerAt", "", "Tracker to Register Server At.")
		registerAs = flag.String("registerAs", "", "Alais for tracker.")
	)
	flag.Parse()

	mel := &dispatcher.Server{
		Me:      *me,
		KeyFile: *keyFile,

		DBConn: *dbConn,
		DBType: *dbType,

		TrackerURL: *registerAt,
		Alias:      *registerAs,
	}
	err := mel.Run(*port)
	if err != nil {
		fmt.Println("Error Starting Server", err)
	}
}
