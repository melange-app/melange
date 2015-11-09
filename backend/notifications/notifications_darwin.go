// +build darwin,!arm

package notifications

import (
	"fmt"

	"github.com/deckarep/gosx-notifier"
)

// Display will push a notification to the platform notification
// manager.
func (n *activeNotification) Display() error {
	note := gosxnotifier.NewNotification(n.Text)
	note.Title = n.Title
	note.Sender = "com.getmelange.Melange"

	// Set the Group to the Sending Plugin
	note.Group = fmt.Sprintf(
		"com.getmelange.Melange.%s",
		n.generatingPlugin.Id,
	)

	return note.Push()
}
