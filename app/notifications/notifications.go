package notifications

import (
	"bytes"
	"fmt"
	"time"

	"text/template"

	"getmelange.com/app/models"
	"getmelange.com/app/packaging"
)

type activeNotification struct {
	Title string
	From  *models.JSONProfile
	Text  string
	Time  time.Time

	generatingPlugin packaging.Plugin
}

// CheckMessageForNotification will determine whether or not a new JSONMessage
// needs to be displayed on screen as a notification.
func CheckMessageForNotification(
	p *packaging.Packager,
	msg *models.JSONMessage,
) (*activeNotification, error) {
	plugs, err := p.AllPlugins()
	if err != nil {
		return nil, err
	}

	for _, v := range plugs {
		for key, n := range v.Notifications {
			if _, ok := msg.Components[key]; ok {
				t, err := template.New("notificationTemplate").Parse(n.Body)
				if err != nil {
					return nil, err
				}

				b := &bytes.Buffer{}
				if err := t.Funcs(map[string]interface{}{
					"component": func(name string) string {
						v, ok := msg.Components[name]
						if !ok {
							return ""
						}
						return v.String
					},
				}).Execute(b, nil); err != nil {
					return nil, err
				}

				title := n.Title
				if title == "" {
					title = fmt.Sprintf(
						"Notification From Melange: %s",
						v.Name,
					)
				}

				return &activeNotification{
					Title:            title,
					Text:             string(b.Bytes()),
					Time:             time.Now(),
					From:             msg.From,
					generatingPlugin: v,
				}, nil
			}
		}
	}

	return nil, nil
}
