package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"airdispat.ch/identity"

	"getmelange.com/zooko/server"
)

func main() {
	var (
		// Namecoin Information
		namecoinServer   = flag.String("namecoin", "", "The URL of the RPC Server.")
		namecoinUsername = flag.String("rpcusername", "", "The username to use when authenticating with the Namecoin RPC Server.")
		namecoinPassword = flag.String("rpcpassword", "", "The password to use when authenticating with the Namecoin RPC Server.")

		// General Server Information
		port        = flag.Int("port", 4763, "The port for the server to listen on.")
		interactive = flag.Bool("i", false, "Whether the server should run in interactive test mode.")
		key_file    = flag.String("key", "", "File to store the server's private keys in.")
		create_key  = flag.Bool("newkey", false, "Create a new key in the file specified.")
	)

	// Go ahead and Parse the Flags
	flag.Parse()

	server := &server.ZookoServer{
		RPCUsername: *namecoinUsername,
		RPCPassword: *namecoinPassword,
		RPCHost:     *namecoinServer,
	}

	if *interactive {
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
			txes, err := server.CreateTransactionListForName(text, false)
			if err != nil {
				fmt.Println("error looking up", err)
				continue
			}

			for _, v := range txes {
				fmt.Println("Found TX", v.TxId)
			}
		}
	} else {
		var (
			loadedKey *identity.Identity
			err       error
		)

		if *key_file == "" || *create_key {
			loadedKey, err = identity.CreateIdentity()
			if err != nil {
				log.Fatal("Unable to create server identity", err)
			}

			if *create_key {
				if err = loadedKey.SaveKeyToFile(*key_file); err != nil {
					log.Fatal("Unable to save key to file", err)
				}
			}
		} else {
			loadedKey, err = identity.LoadKeyFromFile(*key_file)
			if err != nil {
				log.Fatal("Unable to load key from file", err)
			}
		}

		server.Key = loadedKey
		if err := server.Run(*port); err != nil {
			log.Fatal(err)
		}
	}
}
