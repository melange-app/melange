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
				body, err := getTemplate(msg, n.Body)
				if err != nil {
					return nil, err
				}

				title, err := getTemplate(msg, n.Title)
				if err != nil {
					return nil, err
				}

				if title == "" {
					title = fmt.Sprintf(
						"Notification From Melange: %s",
						v.Name,
					)
				}

				return &activeNotification{
					Title:            title,
					Text:             body,
					Time:             time.Now(),
					From:             msg.From,
					generatingPlugin: v,
				}, nil
			}
		}
	}

	return nil, nil
}

func getTemplate(msg *models.JSONMessage, temp string) (string, error) {
	if temp == "" {
		return "", nil
	}

	t, err := template.New("notificationTemplate").Funcs(
		template.FuncMap{
			"component": func(name string) string {
				v, ok := msg.Components[name]
				if !ok {
					return ""
				}
				return v.String
			},
			"from": func() string {
				return msg.From.Name
			},
		}).Parse(temp)
	if err != nil {
		return "", err
	}

	b := &bytes.Buffer{}
	if err := t.Execute(b, nil); err != nil {
		return "", err
	}

	return string(b.Bytes()), nil
}
