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
		dbConn  = flag.String("dbConn", "", "Connection string for Database.")
		dbType  = flag.String("dbType", "", "Type of database. `psql`, `mysql`, `mssql`, `sqlite`.")
		port    = flag.Int("port", 1024, "Port to run server on.")
	)
	flag.Parse()

	mel := &dispatcher.Server{
		Me:      *me,
		KeyFile: *keyFile,
		DBConn:  *dbConn,
		DBType:  *dbType,
	}
	err := mel.Run(*port)
	if err != nil {
		fmt.Println("Error Starting Server", err)
	}
}
