package proof

import (
	"fmt"

	"github.com/melange-app/nmcd/wire"
)

func (p *peer) loadHeaders() error {
	err := p.sendMsgGetHeaders(ResolverLocatorHashes...)
	if err != nil {
		return err
	}

	go func() {
		lastHeight := int32(0)

		for {
			<-p.loadedHeadersChan

			if !p.connected {
				return
			}

			blck, height := p.chainManager.TopHeader()

			if lastHeight == height {
				// We are no longer receiving new headers. We can stop loading.
				fmt.Println("Synchronization complete.")
				p.loadedHeadersChan = nil
				return
			}

			lastHeight = height

			fmt.Println("Attempting to get next headers from", height)
			err := p.sendMsgGetHeaders(&blck.Hash)
			if err != nil {
				fmt.Println("Unable to get headers", err)
			}
		}
	}()

	return nil
}

func (p *peer) sendMsgGetHeaders(from ...*wire.ShaHash) error {
	getHeaders := wire.NewMsgGetHeaders()
	getHeaders.ProtocolVersion = ProtocolVersion

	for _, v := range from {
		getHeaders.AddBlockLocatorHash(v)
	}

	p.writeChan <- getHeaders
	return nil
}

func (p *peer) handleMsgHeaders(w *wire.MsgHeaders) {
	if len(w.Headers) == 0 {
		return
	}

	previousBlock := w.Headers[0]

	blck, _ := p.chainManager.TopHeader()
	checkHash := ResolverLocatorHashes[0]
	if blck != nil {
		checkHash = &blck.Hash
	}

	if !previousBlock.PrevBlock.IsEqual(checkHash) {
		fmt.Println("First header does not match locator hash.")
		return
	}

	var err error
	previousBlock = nil
	for _, current := range w.Headers {
		err = verifiyBlockHeader(previousBlock, current)
		if err != nil {
			fmt.Println("Error verifying incoming header", err)
			return
		}

		previousBlock = current
	}
	p.chainManager.AddHeader(w.Headers...)

	if p.loadedHeadersChan != nil {
		p.loadedHeadersChan <- struct{}{}
	}
}
