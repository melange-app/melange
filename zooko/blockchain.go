package zooko

import (
	"errors"
	"fmt"
	"time"

	"github.com/melange-app/nmcd/wire"
)

type blockHeader struct {
	// We can through out all the auxiliary data once
	// we have verified the block and put it in the chain.
	MerkleRoot wire.ShaHash
	Timestamp  time.Time
	Hash       wire.ShaHash
	Height     int
	// *wire.BlockHeader

	Previous *blockHeader
	Next     *blockHeader
}

type chainBase struct {
	wire.ShaHash
	Height int
	Next   *blockHeader
}

type blockchain struct {
	Base *chainBase
	Top  *blockHeader

	Direct map[string]*blockHeader

	Height int
}

// will only accept headers if they are the "next in the chain"
func (b *blockchain) addHeader(h *wire.BlockHeader) error {
	hash, err := h.BlockSha()
	if err != nil {
		return err
	}

	if _, ok := b.Direct[hash.String()]; ok {
		return errors.New("We already have this block...")
	}

	// Is the PreviousBlock Hash already in the Chain?
	if check, ok := b.Direct[h.PrevBlock.String()]; ok {
		if check != b.Top {
			// Since this block is not already in the chain and it isn't
			// appending, then we have some sort of fork.
			return fmt.Errorf(
				"Rejecting Block Header (%s) because of fork at (%d)",
				hash.String()[:20],
				check.Height,
			)
		}
	} else if b.Top == nil {
		// This block could be at the top
		if !b.Base.IsEqual(&h.PrevBlock) {
			return fmt.Errorf(
				"PrevBlcok Hash (%s) does not match top of chain (%s).",
				h.PrevBlock.String()[:20],
				b.Base.String()[:20],
			)
		}

	}

	// Create the Header
	hdr := &blockHeader{
		MerkleRoot: h.MerkleRoot,
		Timestamp:  h.Timestamp,
		Hash:       hash,
		Height:     b.Height + b.Base.Height + 1,
	}
	b.Direct[hash.String()] = hdr

	if b.Top == nil {
		// We just add this directly to the chain base.
		b.Base.Next = hdr
		b.Top = hdr
	} else {
		// We will add on to the the top of the chain
		b.Top.Next = hdr
		hdr.Previous = b.Top
		b.Top = hdr
	}
	b.Height++

	fmt.Println("ACCEPTED Header", hash.String()[:20], "at height", b.height())

	return nil
}

func (b *blockchain) height() int32 {
	return int32(b.Height) + int32(b.Base.Height)
}

type blockchainManager struct {
	acceptChannel chan interface{}
	chain         *blockchain
}

func (b *blockchainManager) TopHeader() (*blockHeader, int32) {
	return b.chain.Top, b.chain.height()
}

func CreateBlockchainManager() *blockchainManager {
	b := &blockchainManager{
		acceptChannel: make(chan interface{}),
		chain: &blockchain{
			Base: &chainBase{
				ShaHash: *ResolverLocatorHashes[0],
				Height:  TopResolverHeight,
			},
			Direct: make(map[string]*blockHeader),
		},
	}
	go b.acceptanceLoop()

	return b
}

func (b *blockchainManager) acceptanceLoop() {
	for {
		obj := <-b.acceptChannel

		switch t := obj.(type) {
		case *wire.BlockHeader:
			b.addHeader(t)
		}
	}
}

func (b *blockchainManager) addHeader(t *wire.BlockHeader) {
	err := b.chain.addHeader(t)
	if err != nil {
		fmt.Println("Unable to add header to chain", err)
	}
}

func (b *blockchainManager) AddHeader(hdrs ...*wire.BlockHeader) {
	// fmt.Println(hdrs[0].PrevBlock)
	for _, v := range hdrs {
		b.acceptChannel <- v
	}
}
