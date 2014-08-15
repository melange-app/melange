package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// this will actually install the update
func main() {
	var (
		app = flag.String("app", "", "Location of the new application to launch.")
		old = flag.String("old", "", "Location of the old application to killl.")
	)
	flag.Parse()

	time.Sleep(500 * time.Millisecond)

	// Remove old Melange
	err := os.RemoveAll(*old)
	if err != nil {
		fmt.Println("Can't remove old melange.", err)
	}

	err = exec.Command("open", *app).Start()
	if err != nil {
		fmt.Println("Error starting Melange again.")
	}
}
