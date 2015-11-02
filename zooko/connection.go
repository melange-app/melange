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
