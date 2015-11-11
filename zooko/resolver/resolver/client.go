package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"getmelange.com/zooko/account"
	"getmelange.com/zooko/resolver"
)

type handler struct {
	Client  *resolver.Client
	Account *account.Account
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	h := &handler{
		Client: resolver.CreateClient(nil),
	}

	for {
		fmt.Print("> ")

		// Get the command from stdin
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "quit" {
			return
		}

		commands := strings.Split(text, " ")
		if commands[0] == "lookup" && len(commands) == 2 {
			h.lookup(commands[1])
		} else if commands[0] == "register" && len(commands) == 3 {
			h.register(commands[1], commands[2])
		} else if commands[0] == "generate_key" {
			h.generateKey()
		} else if commands[0] == "load_key" && len(commands) == 2 {
			h.loadKey(commands[1])
		} else if commands[0] == "save_key" && len(commands) == 2 {
			h.saveKey(commands[1])
		} else {
			fmt.Println("Unrecognized command")
		}
	}
}

func (h *handler) printKey() {
	hash, err := h.Account.PublicKeyHash()
	if err != nil {
		fmt.Println("couldn't print out namecoin address")
		return
	}

	fmt.Println("Has address", hash)
}

func (h *handler) generateKey() {
	acc, err := account.CreateAccount()
	if err != nil {
		fmt.Println("error creating namecoin account", err)
		return
	}

	h.Account = acc
	h.printKey()
}

func (h *handler) loadKey(file string) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Println("error loading namecoin key", err)
		return
	}
	defer f.Close()

	acc, err := account.CreateAccountFromReader(f)
	if err != nil {
		fmt.Println("error reading namecoin key", err)
		return
	}

	h.Account = acc
	h.printKey()
}

func (h *handler) saveKey(file string) {
	f, err := os.Create(file)
	if err != nil {
		fmt.Println("error saving namecoin key", err)
		return
	}
	defer f.Close()

	h.Account.Serialize(f)
}

func (h *handler) register(name, value string) {
	if h.Account == nil {
		fmt.Println("Load or generate namecoin key first")
	}

	err := h.Client.Register(name, &resolver.Registration{
		Location: value,
	}, h.Account)
	if err != nil {
		fmt.Println("error registering name", err)
	} else {
		fmt.Println("successfully registered")
	}
}

func (h *handler) lookup(text string) {
	fmt.Println("Looking up", text)
	val, found, err := h.Client.Lookup(text)
	if err != nil {
		fmt.Println("error looking up", err)
		return
	}

	if !found {
		fmt.Println("That name does not exist.")
		return
	}

	fmt.Println("Value for", text, "is", val)
}
