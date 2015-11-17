package dev

type Command struct{}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Run(options []string) {

}

func (c *Command) Description() string {
	return "package Melange applications"
}

func (c *Command) Help() string {
	return `usage: melange dev [command]

Dev will handle the packaging of Melange applications for distribution.
`
}
