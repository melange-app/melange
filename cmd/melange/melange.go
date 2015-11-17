package main

import (
	"fmt"
	"os"

	"getmelange.com/cmd/dev"
	"getmelange.com/cmd/nmc"
	"getmelange.com/cmd/srv"
)

type Runner interface {
	Run(options []string)
	Description() string
	Help() string
}

var commands = map[string]Runner{
	"dev": dev.NewCommand(),
	"srv": srv.NewCommand(),
	"nmc": nmc.NewCommand(),
}
var sortedCommands = []string{"dev", "nmc", "srv"}

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		printHelp("")
		return
	}

	command := arguments[1]
	options := arguments[2:]

	if command == "help" {
		topic := ""
		if len(options) == 1 {
			topic = options[0]
		}

		printHelp(topic)
		return
	}

	runner, ok := commands[command]
	if !ok {
		printNotFound(command)
		return
	}

	runner.Run(options)
}

var (
	description = `Melange is a tool for interacting with Melange resources.

Usage:

	melange command [arguments]

The commands are:
`

	suffix = `
Use "melange help [command] for more information about a command."
`
)

func printHelp(cmd string) {
	if cmd == "" {
		fmt.Println(description)

		for _, name := range sortedCommands {
			runner := commands[name]
			if runner == nil {
				continue
			}

			fmt.Printf("\t%s\t%s\n", name, runner.Description())
		}

		fmt.Println(suffix)
		return
	}

	runner, ok := commands[cmd]
	if !ok {
		fmt.Printf("Unknown help topic `%s`. Run 'melange help'.\n", cmd)
		return
	}

	fmt.Println(runner.Help())
}

func printNotFound(cmd string) {
	fmt.Printf("melange: unknown subcommand `%s`\n", cmd)
	fmt.Printf("Run 'melange help' for usage.\n")
}
