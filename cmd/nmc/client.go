package nmc

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"getmelange.com/zooko/account"
	"getmelange.com/zooko/resolver"
)

type Command struct {
	Client  *resolver.Client
	Account *account.Account
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Description() string {
	return "provide low-level access to Namecoin resources"
}

func (c *Command) Help() string {
	return `usage: melange nmc [command]`
}

func (h *Command) Run(options []string) {
	reader := bufio.NewReader(os.Stdin)
	h.Client = resolver.CreateClient(nil)

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
			h.lookup(commands[1], false)
		} else if commands[0] == "lookup" && len(commands) == 3 {
			h.lookup(commands[1], true)
		} else if commands[0] == "register" && len(commands) == 3 {
			h.register(commands[1], commands[2])
		} else if commands[0] == "generate_key" {
			h.generateKey()
		} else if commands[0] == "load_key" && len(commands) == 2 {
			h.loadKey(commands[1])
		} else if commands[0] == "save_key" && len(commands) == 2 {
			h.saveKey(commands[1])
		} else if commands[0] == "addtx" && len(commands) == 5 {
			h.addTransaction(commands[1], commands[2], commands[3], commands[4])
		} else if commands[0] == "balance" {
			h.balance()
		} else if commands[0] == "getutxo" {
			h.utxo()
		} else if commands[0] == "export" {
			h.export()
		} else {
			fmt.Println("Unrecognized command")
		}
	}
}

func (h *Command) export() {
	data := h.Account.Keys.Key().Serialize()
	fmt.Println("Namecoin private key")
	fmt.Println(hex.EncodeToString(data))
}

func (h *Command) printKey() {
	hash, err := h.Account.PublicKeyHash()
	if err != nil {
		fmt.Println("couldn't print out namecoin address")
		return
	}

	fmt.Println("Has address", hash)
}

func (h *Command) generateKey() {
	acc, err := account.CreateAccount()
	if err != nil {
		fmt.Println("error creating namecoin account", err)
		return
	}

	h.Account = acc
	h.printKey()
}

func (h *Command) loadKey(file string) {
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

func (h *Command) saveKey(file string) {
	f, err := os.Create(file)
	if err != nil {
		fmt.Println("error saving namecoin key", err)
		return
	}
	defer f.Close()

	h.Account.Serialize(f)
}

func (h *Command) register(name, value string) {
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

func (h *Command) lookup(text string, prefix bool) {
	if !prefix {
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
		return
	}

	fmt.Println("Looking up", text)
	val, err := h.Client.LookupPrefix(text)
	if err != nil {
		fmt.Println("error looking up", err)
		return
	}
	fmt.Println("Found", len(val), "names.")
	fmt.Println(val)

	return
}

func (h *Command) balance() {
	fmt.Println("account balance", h.Account.Balance())
	fmt.Println("name minimum", int64(account.NameMinimumBalance))
}

func (h *Command) utxo() {
	fmt.Println("printing unspent")
	for _, v := range h.Account.Unspent {
		fmt.Println(v.TxID)
	}
}

func (h *Command) addTransaction(txid, output, amount, pkscript string) {

}
