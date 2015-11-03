package proof

import (
	"fmt"
	"io"
	"net"

	"github.com/melange-app/nmcd/wire"
)

// ConnectToPeer will open a connection to a peer's address.
func connectToPeer(id string) (net.Conn, error) {
	toConnect := fmt.Sprintf("%s:%d", id, PeeringPort)
	conn, err := net.Dial("tcp", toConnect)
	if err != nil {
		return nil, err
	}

	fmt.Println("Successfully connected to peer", id)

	err = sendVersion(conn)
	if err != nil {
		return nil, err
	}

	// Load the Remote Node's Protocol Version
	msg, _, err := wire.ReadMessage(conn, ProtocolVersion, NamecoinNet)
	if err != nil {
		return nil, err
	}

	remoteVersion, ok := msg.(*wire.MsgVersion)
	if !ok {
		return nil, fmt.Errorf("Incorrect message returned %s", msg.Command())
	}
	fmt.Println(
		"Connected to",
		id,
		"with version",
		remoteVersion.ProtocolVersion,
		"last block",
		remoteVersion.LastBlock,
	)

	// Set the ProtocolVersion to the minimum of the two clients.
	if uint32(remoteVersion.ProtocolVersion) < ProtocolVersion {
		ProtocolVersion = uint32(remoteVersion.ProtocolVersion)
	}

	// Get VerAck from Remote Peer
	msg, _, err = wire.ReadMessage(conn, ProtocolVersion, NamecoinNet)
	if err != nil {
		return nil, err
	}

	_, ok = msg.(*wire.MsgVerAck)
	if !ok {
		return nil, incorrectCommand(msg)
	}

	return conn, nil
}

func sendVersion(conn net.Conn) error {
	nonce++

	ver, err := wire.NewMsgVersionFromConn(conn, nonce, int32(TopResolverHeight))
	if err != nil {
		return err
	}

	return wire.WriteMessage(conn, ver, ProtocolVersion, NamecoinNet)
}

type peer struct {
	conn              net.Conn
	writeChan         chan wire.Message
	loadedHeadersChan chan struct{}
	connected         bool
	chainManager      *blockchainManager
}

func newPeer(id string, b *blockchainManager) (*peer, error) {
	conn, err := connectToPeer(id)
	if err != nil {
		return nil, err
	}

	p := &peer{
		conn:              conn,
		writeChan:         make(chan wire.Message),
		loadedHeadersChan: make(chan struct{}),
		connected:         true,
		chainManager:      b,
	}

	go p.readLoop()
	go p.writeLoop()

	return p, nil
}

func (p *peer) readLoop() {
	for {
		msg, _, err := wire.ReadMessage(p.conn, ProtocolVersion, NamecoinNet)

		if err == io.EOF {
			p.connected = false
			fmt.Println("Disconnected from Peer", p)
			return
		}

		if err != nil {
			fmt.Println("Got error reading message", err)
			continue
		}

		fmt.Println("Got message", msg.Command())

		switch t := msg.(type) {
		case *wire.MsgHeaders:
			p.handleMsgHeaders(t)
		case *wire.MsgInv:
			p.handleMsgInv(t)
		case *wire.MsgAddr:
			p.handleMsgAddr(t)
		}
	}
}

func (p *peer) writeLoop() {
	for {
		msg := <-p.writeChan
		if !p.connected {
			return
		}

		err := wire.WriteMessage(p.conn, msg, ProtocolVersion, NamecoinNet)
		if err != nil {
			fmt.Println("Got error writing message", err)
			continue
		}
		fmt.Println("Wrote message", msg.Command())
	}
}
