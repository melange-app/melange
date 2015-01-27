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
	"time"

	"getmelange.com/app/packaging"
)

func init() {
	go func() {
		<-time.After(10 * time.Second)
		fmt.Println("Displaying Notification")
		(&activeNotification{
			Title: "Hello, Go Notifications",
			Text:  "Blah blah",
			generatingPlugin: packaging.Plugin{
				Id: "testPlugin",
			},
		}).Display()
	}()
}

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
