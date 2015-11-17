package srv

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	dispatcherPrivateKey = "dispatcher.pk"
	namecoinAccount      = "namecoin.nmc"

	serverName        = "server.name"
	serverDescription = "server.desc"

	serverIdentifier = "server.identifier"
	serverLocation   = "server.location"
)

type PublishRequest struct {
	Reader *bufio.Reader
}

func createPublishRequest() *PublishRequest {
	return &PublishRequest{
		Reader: bufio.NewReader(os.Stdin),
	}
}

func (p *PublishRequest) getInput(name string) error {
	text, _ := p.Reader.ReadString('\n')
	text = strings.TrimSpace(text)

	return nil
}

func (c *Command) RunPublish(extra []string) {
	p := createPublishRequest()

	// Overview
	fmt.Println("This tool will help you publish a Dispatcher server to the blockchain for discover.")
	fmt.Println("Please answer each of the following questions:")

	// Server Key Information
	fmt.Println("These questions deal with the server keys.")

	fmt.Print("Dispatcher Private Key Location [*.pk]: ")
	if err := p.getInput(dispatcherPrivateKey); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Print("Namecoin Account Location [*.nmc]: ")
	if err := p.getInput(namecoinAccount); err != nil {
		fmt.Println(err)
		return
	}

	// Description Information
	fmt.Println("These questions refer to the human-readable listing of the server in the directory.")

	fmt.Print("Server Name [Melange Official Server]: ")
	if err := p.getInput(serverName); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Print("Server Description: ")
	if err := p.getInput(serverDescription); err != nil {
		fmt.Println(err)
		return
	}

	// Location information
	fmt.Println("These questions refer to the information that Melange uses to locate the server.")

	fmt.Print("Server Identifier (must be unique) [adme]: ")
	if err := p.getInput(serverIdentifier); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Print("Server Location [mailserver.airdispatch.me:2048]: ")
	if err := p.getInput(serverLocation); err != nil {
		fmt.Println(err)
		return
	}
}
