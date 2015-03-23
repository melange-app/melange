package zooko

import (
	"fmt"

	"github.com/melange-app/nmcd/wire"
)

func incorrectCommand(actual wire.Message) error {
	return fmt.Errorf("Incorrect message type. Received %s.", actual.Command())
}

func shaHashFromString(str string) *wire.ShaHash {
	hash, err := wire.NewShaHashFromStr(str)
	if err != nil {
		panic(err)
	}

	return hash
}

var nonce uint64
var quitChan = make(chan bool)

// ConnectToNetwork will attempt to connect to the bootstrapping peers.
func ConnectToNetwork() error {
	chainMananger := CreateBlockchainManager()

	for _, v := range BootstrapNodes {
		p, err := newPeer(v, chainMananger)
		if err != nil {
			fmt.Println("Got error connecting to peer", err)
		}

		if err = p.loadHeaders(); err != nil {
			fmt.Println("Got error getting headers", err)
		}
	}

	fmt.Println("Waiting to Quit")
	<-quitChan
	return nil
}
