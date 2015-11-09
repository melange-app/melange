// +build android

package notifications

/*
#cgo LDFLAGS: -llog -landroid

void createNotification(char* title, char* body, char* id);
*/
import "C"
import (
	"fmt"
	"runtime"
)

var notificationChannel = make(chan *activeNotification)

//export createNotificationThread
func createNotificationThread() {
	runtime.LockOSThread()

	for {
		note := <-notificationChannel
		note.displayThread()
	}

	// runtime.UnlockOSThread()
}

func (n *activeNotification) Display() error {
	if notificationChannel == nil {
		return nil
	}

	notificationChannel <- n
	return nil
}

func (n *activeNotification) displayThread() error {
	cTitle := C.CString(n.Title)
	cBody := C.CString(n.Text)
	cId := C.CString(fmt.Sprintf("%s:%s", n.generatingPlugin.Id, n.Title))

	_, err := C.createNotification(cTitle, cBody, cId)
	return err
}
