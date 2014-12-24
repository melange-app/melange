// +build android

package framework

import (
	"fmt"
	"io"
	"path/filepath"

	"getmelange.com/app/replacer"
	"golang.org/x/mobile/app"
)

func GetFile(prefix, request string) (*FileView, error) {
	fmt.Println("Opening file", prefix, request)
	if prefix == "client" || prefix == "lib" {
		var fileReader io.ReadCloser

		// fmt.Println("About to App.Open")
		fileReader, err := app.Open(filepath.Join(prefix, request))

		// fmt.Println("Returned from app.Open")
		if err != nil {
			fmt.Println("Got App Open Error", err)
			return nil, err
		}
		// fmt.Println("Just App Opened", filepath.Ext(request))

		switch filepath.Ext(request) {
		case ".html":
			fallthrough
		case ".js":
			fmt.Println("Replacing with the File Reader", request)
			fileReader = replacer.CreateReplacer(
				fileReader,
				`http://([a-z\.]*).melange(:7776)?`,
				`http://$1.melange.127.0.0.1.xip.io:7776`,
				`[^a-z\.]`,
			)
		}

		components := filepath.SplitList(request)

		// fmt.Println("Successfully got file")

		return &FileView{
			File: fileReader,
			Name: components[len(components)-1],
			Size: -1,
		}, nil
	} else if prefix == "plugins" {
		// Pretty much every other
	}

	return nil, errNoFile
}
