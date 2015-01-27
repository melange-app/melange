// +build !android,!darwin

package notifications

func (n *activeNotification) Display() {
	// Other platforms do not have notifications built in.
	return nil
}
