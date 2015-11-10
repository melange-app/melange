package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"getmelange.com/zooko/resolver"
)

func main() {
	client := resolver.CreateClient(nil)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		// Get the command from stdin
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "quit" {
			return
		}

		fmt.Println("Looking up", text)
		val, found, err := client.Lookup(text)
		if err != nil {
			fmt.Println("error looking up", err)
			continue
		}

		if !found {
			fmt.Println("That name does not exist.")
			continue
		}

		fmt.Println("Value for", text, "is", string(val))
	}
}
