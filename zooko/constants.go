package zooko

import (
	"github.com/melange-app/nmcd/wire"
)

// bootstrap nodes: 69.164.206.88 and 78.47.86.43
// port 8333 for peer to peer

const (
	// NamecoinNet is the magic number for the Namecoin network.
	NamecoinNet wire.BitcoinNet = 0xfeb4bef9

	// PeeringPort is the port to connect to peers using.
	PeeringPort = 8334
)

var (
	// ResolverLocatorHashes is the hash of the earliest block (at the time of development)
	// that could possibly contain an active name. Currently, this is set to the hash of block
	// at height 181067. This is very unfortunate because merged mining only begins at block
	// 19200.
	ResolverLocatorHashes = []*wire.ShaHash{
		shaHashFromString("5c157307de4f94cb76a2853259df25a118ebacaf3c6a597d512aa46fddb459f9"),
		// shaHashFromString("7bc9fceb1e8b502421fac237d700c6dcc357144f5fc5e0d8f23b63b306266373"),
		// shaHashFromString("23c8df0e5f465be03c23342d50b4c491561517237ae03f706f76887feed45545"),
		// shaHashFromString("5a37a2258a09737b01a12a33a69dffce8f5cff334c37c6898d0624863254f312"),
		// shaHashFromString("a37f87eab6a35037b96f13155a7e54a6171089d4a6f50b7039592e31660a6e69"),
		// shaHashFromString("52526049f1ddbfb778f3bdab2e8fcfffc2f48c2ff10a2f3e764f904498555804"),
		shaHashFromString("000000000036cba77f97fb033dd770830614e2e4f21b61abb0a893472d0bbbb5"),
	}

	// TopResolverHeight is the height of the last known block this library was
	// written for.
	TopResolverHeight = 181067
	// TopResolverHeight = 19412
	// TopResolverHeight = 39424
	// TopResolverHeight = 139311
	// TopResolverHeight = 138785
	// TopResolverHeight = 216116

	// ProtocolVersion is the version of the Namecoin protocol that this client supports.
	ProtocolVersion uint32 = 38000

	// BootstrapNodes contains the IP addresses of the Namecoin
	// nodes to connect to first.
	BootstrapNodes = []string{
		// "69.164.206.88", <-- This node always gives me a timeout. :/
		// "78.47.86.43",
		"localhost",
	}

	// BootstrapNodes = []string{"localhost"}

)
