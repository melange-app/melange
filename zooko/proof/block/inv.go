package proof

import (
	"fmt"

	"github.com/melange-app/nmcd/wire"
)

func (p *peer) handleMsgInv(t *wire.MsgInv) {
	for _, v := range t.InvList {
		fmt.Println("Received Inv", v.Type, v.Hash.String()[:20])
		if v.Type == wire.InvTypeBlock {
			// Oh, snap! A new block is available! Let's kick off a new request.
			block, _ := p.chainManager.TopHeader()
			err := p.sendMsgGetHeaders(&block.Hash)
			if err != nil {
				fmt.Println("Error getting new message headers", err)
			} else {
				return
			}
		}
	}
}
