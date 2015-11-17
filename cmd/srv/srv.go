package srv

import "fmt"

type Command struct{}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Run(options []string) {
	if len(options) == 0 {
		fmt.Println(c.Help())
		return
	}

	command := options[0]
	extra := options[1:]

	switch command {
	case "publish":
		c.RunPublish(extra)
	}
}

func (c *Command) Description() string {
	return "publish Dispatcher servers"
}

func (c *Command) Help() string {
	return `usage: melange srv [command]

Server will handle the publishing of Dispatcher servers for users.

List of commands:

	publish		walk through publishing a server to Melange
`
}
