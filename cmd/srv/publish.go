package srv

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"airdispat.ch/identity"
	"getmelange.com/zooko/account"
	"getmelange.com/zooko/resolver"
)

const (
	dispatcherPrivateKey = "dispatcher.pk"
	namecoinAccount      = "namecoin.nmc"

	serverName        = "server.name"
	serverDescription = "server.desc"

	serverIdentifier = "server.identifier"
	serverLocation   = "server.location"

	serverPrefix = "mlg/server/"

	confirmation = "[y/n]"
)

var (
	errEmpty           = errors.New("This field is required and cannot be blank.")
	errNoExist         = errors.New("The file you specified does not exist.")
	errParseIssue      = errors.New("Cannot read specified file. Perhaps it is corrupted?")
	errIDTaken         = errors.New("That identifier is already taken, please choose another.")
	errNo              = errors.New("errno")
	errNotConfirmation = errors.New("Please say either 'y' for yes or 'n' for no.")
)

type PublishRequest struct {
	reader   *bufio.Reader
	identity *identity.Identity
	client   *resolver.Client

	account     *account.Account
	accountFile string

	ID       string
	Location string
}

func createPublishRequest() *PublishRequest {
	return &PublishRequest{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (p *PublishRequest) checkEmpty(text string) error {
	if text == "" {
		return errEmpty
	}

	return nil
}

func (p *PublishRequest) checkExists(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return errNoExist
	}

	return nil
}

func (p *PublishRequest) checkNamecoinAccount(file string) error {
	if err := p.checkEmpty(file); err != nil {
		return err
	}
	if err := p.checkExists(file); err != nil {
		return err
	}

	p.accountFile = file
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	acc, err := account.CreateAccountFromReader(f)
	if err != nil {
		return errParseIssue
	}

	p.account = acc
	return nil
}

func (p *PublishRequest) checkDispatcherKey(file string) error {
	if err := p.checkEmpty(file); err != nil {
		return err
	}
	if err := p.checkExists(file); err != nil {
		return err
	}

	loadedKey, err := identity.LoadKeyFromFile(file)
	if err != nil {
		return errParseIssue
	}

	p.identity = loadedKey
	p.client = resolver.CreateClient(p.identity)
	return nil
}

func (p *PublishRequest) checkServerIdentifier(id string) error {
	if err := p.checkEmpty(id); err != nil {
		return err
	}

	if _, found, err := p.client.Lookup(serverPrefix + id); err != nil {
		return err
	} else if found {
		return errIDTaken
	}

	p.ID = id
	return nil
}

func (p *PublishRequest) getInput(name string) error {
	text, _ := p.reader.ReadString('\n')
	text = strings.TrimSpace(text)

	switch name {
	case serverLocation:
		p.Location = text
		return p.checkEmpty(text)

	case serverIdentifier:
		return p.checkServerIdentifier(text)

	case dispatcherPrivateKey:
		return p.checkDispatcherKey(text)
	case namecoinAccount:
		return p.checkNamecoinAccount(text)

	case confirmation:
		if text == "y" || text == "Y" || text == "yes" {
			return nil
		} else if text == "n" || text == "N" || text == "no" {
			return errNo
		} else {
			return errNotConfirmation
		}
	}

	return nil
}

func (p *PublishRequest) performRegistration() error {
	reg := resolver.CreateRegistrationFromIdentity(
		p.identity, // Identity
		p.ID,       // Identifier
		p.Location, // Location
	)

	if err := p.client.Register(serverPrefix+p.ID, reg, p.account); err != nil {
		return err
	}

	file, err := os.Create(p.accountFile)
	if err != nil {
		return err
	}
	defer file.Close()

	if err = p.account.Serialize(file); err != nil {
		return err
	}

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

	for {
		fmt.Print("Proceed with registration? [y/n]: ")
		if err := p.getInput(confirmation); err != nil {
			if err == errNo {
				fmt.Println("Registration aborted.")
				return
			}

			fmt.Println(err)
			continue
		}
		break
	}

	if err := p.performRegistration(); err != nil {
		fmt.Println("Encountered an error during registration.")
		fmt.Println(err)
		return
	}

	fmt.Println("The server has been successfully registered for use with Melange. ")
	fmt.Println("It can take 2 - 3 hours for a registration to appear on the blockchain, until that time we will serve your registration from an intermediate server.")
}
