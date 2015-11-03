package proof

import (
	"fmt"

	"github.com/melange-app/nmcd/wire"
)

func (p *peer) handleMsgAddr(t *wire.MsgAddr) {
	fmt.Println("Peer", p, "sends", len(t.AddrList), "addresses.")
}
