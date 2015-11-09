package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework UIKit -framework Foundation
#import "AppDelegate.h"
#import <UIKit/UIKit.h>

extern int StartApp(int argc, char * argv []);
*/
import "C"
import (
	"fmt"
	"os"

	"getmelange.com/mobile/melange"
)

func main() {

	argc := C.int(len(os.Args))
	argv := make([]*C.char, argc)
	for i, arg := range os.Args {
		argv[i] = C.CString(arg)
	}

	C.StartApp(argc, &argv[0])
}

//export startGoServer
func startGoServer(dataPath *C.char, assetsPath *C.char) {
	err := melange.RunDarwin(
		7776,
		C.GoString(dataPath),
		C.GoString(assetsPath),
		"0.1",
		"ios",
	)
	if err != nil {
		fmt.Println("iOS Server has Crashed:", err)
		return
	}
}

//export hasNewContent
func hasNewContent() bool {
	return false
}
